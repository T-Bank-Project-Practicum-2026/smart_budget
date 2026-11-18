package tbankcsv

import (
	"bufio"
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
	"time"
	"unicode"
)

const (
	columnAccount    = "Имя счёта"
	columnCard       = "Номер карты"
	columnDate       = "Дата операции"
	columnAmount     = "Сумма в валюте счёта"
	columnCurrency   = "Валюта счёта"
	columnStatus     = "Статус"
	columnCategory   = "Категория по-умолчанию"
	columnMCC        = "MCC"
	columnMerchant   = "Описание"
	columnAnalytics  = "Учёт в аналитике"
	dateLayout       = "02.01.2006 15:04:05"
	currencyRUB      = "RUB"
	maxRowErrors     = 100
	utf8BOM          = "\xef\xbb\xbf"
	keySeparator     = "\x1f"
	sourceKeyPrefix  = "tbank_csv:"
	sourceKeyName    = "tbank_csv"
	statusOK         = "Ок"
	analyticsYes     = "Да"
	analyticsNo      = "Нет"
	categoryTransfer = "переводы"
	categoryDeposit  = "пополнения"
	categoryBonus    = "бонусы"
	categoryInterest = "проценты"
)

var (
	// В CSV нет даже строки заголовка
	ErrEmptyFile = errors.New("empty csv file")

	// Заголовок не содержит обязательных колонок
	ErrInvalidHeader = errors.New("invalid csv header")

	// CSV структурно повреждён
	ErrInvalidCSV = errors.New("invalid csv")

	// Фиксируем московское время
	moscowTimeZone = time.FixedZone("MSK", 3*60*60)
)

// Row — нормализованная операция-расход из выписки.
type Row struct {
	Line           int
	AccountName    string
	CardMask       string
	OccurredAt     time.Time
	AmountKopecks  int64
	Currency       string
	Merchant       string
	MCC            string
	SourceCategory string
	ExternalID     string
	RowHash        string
}

type RowError struct {
	Line    int    `json:"line"`
	Message string `json:"message"`
}

type Result struct {
	TotalRows       int
	Rows            []Row
	SkippedRows     int
	FailedRows      int
	Errors          []RowError
	ErrorsTruncated bool
}

// Parse читает CSV-выписку T-Bank. В Rows попадают только расходы, которые
// должны уменьшать бюджет. Доходы, переводы и служебные операции считаются
// пропущенными, а некорректные строки — ошибками, не прерывающими импорт.
func Parse(input io.Reader) (Result, error) {
	reader := csv.NewReader(stripBOM(input))
	reader.Comma = ';'
	reader.FieldsPerRecord = -1

	header, err := reader.Read()
	if errors.Is(err, io.EOF) {
		return Result{}, ErrEmptyFile
	}

	if err != nil {
		return Result{}, fmt.Errorf("%w: %w", ErrInvalidCSV, err)
	}

	columns, err := mapHeader(header)
	if err != nil {
		return Result{}, err
	}

	reader.FieldsPerRecord = len(header)
	result := Result{}

	for {
		record, err := reader.Read()
		if errors.Is(err, io.EOF) {
			break
		}

		if errors.Is(err, csv.ErrFieldCount) {
			result.TotalRows++
			result.FailedRows++
			result.addError(recordLine(reader, record), "неверное число колонок")

			continue
		}

		if err != nil {
			return Result{}, fmt.Errorf("%w: %w", ErrInvalidCSV, err)
		}

		result.TotalRows++

		line := recordLine(reader, record)
		row, skipped, err := parseRecord(record, columns)
		if err != nil {
			result.FailedRows++
			result.addError(line, err.Error())

			continue
		}

		if skipped {
			result.SkippedRows++

			continue
		}

		row.Line = line
		result.Rows = append(result.Rows, row)
	}

	return result, nil
}

type columns struct {
	account   int
	card      int
	date      int
	amount    int
	currency  int
	status    int
	category  int
	mcc       int
	merchant  int
	analytics int
}

func mapHeader(record []string) (columns, error) {
	index := make(map[string]int, len(record))

	for position, value := range record {
		name := normalizeSpace(value)
		if name == "" {
			continue
		}

		if _, exists := index[name]; exists {
			return columns{}, fmt.Errorf("%w: duplicate column %q", ErrInvalidHeader, name)
		}

		index[name] = position
	}

	result := columns{}

	for _, item := range []struct {
		name   string
		target *int
	}{
		{columnAccount, &result.account},
		{columnCard, &result.card},
		{columnDate, &result.date},
		{columnAmount, &result.amount},
		{columnCurrency, &result.currency},
		{columnStatus, &result.status},
		{columnCategory, &result.category},
		{columnMCC, &result.mcc},
		{columnMerchant, &result.merchant},
		{columnAnalytics, &result.analytics},
	} {
		position, exists := index[item.name]
		if !exists {
			return columns{}, fmt.Errorf("%w: missing column %q", ErrInvalidHeader, item.name)
		}

		*item.target = position
	}

	return result, nil
}

func parseRecord(record []string, columns columns) (Row, bool, error) {
	status := normalizeSpace(field(record, columns.status))

	switch {
	case status == "":
		return Row{}, false, errors.New("пустой статус операции")

	case !strings.EqualFold(status, statusOK):
		return Row{}, true, nil
	}

	analytics := normalizeSpace(field(record, columns.analytics))

	switch {
	case analytics == "":
		return Row{}, false, errors.New("пустой признак «Учёт в аналитике»")

	case strings.EqualFold(analytics, analyticsNo):
		return Row{}, true, nil

	case !strings.EqualFold(analytics, analyticsYes):
		return Row{}, false, errors.New("неверный признак «Учёт в аналитике»")
	}

	category := normalizeSpace(field(record, columns.category))
	if skipCategory(category) {
		return Row{}, true, nil
	}

	amount, err := parseAmount(field(record, columns.amount))
	if err != nil {
		return Row{}, false, errors.New("неверная сумма в валюте счёта")
	}

	if amount >= 0 {
		return Row{}, true, nil
	}

	account := strings.TrimSpace(field(record, columns.account))
	if account == "" {
		return Row{}, false, errors.New("пустое имя счёта")
	}

	occurredAt, err := time.ParseInLocation(dateLayout, strings.TrimSpace(field(record, columns.date)), moscowTimeZone)
	if err != nil {
		return Row{}, false, errors.New("неверная дата операции")
	}

	currency := strings.ToUpper(strings.TrimSpace(field(record, columns.currency)))
	if currency != currencyRUB {
		return Row{}, false, errors.New("неподдерживаемая валюта счёта")
	}

	mcc := strings.TrimSpace(field(record, columns.mcc))
	if mcc != "" && !isMCC(mcc) {
		return Row{}, false, errors.New("неверный MCC")
	}

	row := Row{
		AccountName:    account,
		CardMask:       strings.TrimSpace(field(record, columns.card)),
		OccurredAt:     occurredAt.UTC(),
		AmountKopecks:  amount,
		Currency:       currency,
		Merchant:       strings.TrimSpace(field(record, columns.merchant)),
		MCC:            mcc,
		SourceCategory: category,
	}
	row.ExternalID, row.RowHash = deriveKeys(row)

	return row, false, nil
}

func skipCategory(category string) bool {
	switch strings.ToLower(category) {
	case categoryTransfer, categoryDeposit, categoryBonus, categoryInterest:
		return true

	default:
		return false
	}
}

func deriveKeys(row Row) (string, string) {
	identity := strings.Join([]string{
		sourceKeyName,
		row.AccountName,
		row.CardMask,
		row.OccurredAt.UTC().Format(time.RFC3339Nano),
		strconv.FormatInt(row.AmountKopecks, 10),
		row.Currency,
	}, keySeparator)

	content := strings.Join([]string{
		identity,
		row.Merchant,
		row.MCC,
	}, keySeparator)

	return sourceKeyPrefix + hashString(identity), hashString(content)
}

func hashString(value string) string {
	sum := sha256.Sum256([]byte(value))

	return hex.EncodeToString(sum[:])
}

func parseAmount(value string) (int64, error) {
	value = strings.TrimSpace(value)
	value = strings.ReplaceAll(value, "\u2212", "-")
	value = strings.Map(func(char rune) rune {
		if unicode.IsSpace(char) {
			return -1
		}

		return char
	}, value)

	if value == "" {
		return 0, errors.New("empty amount")
	}

	negative := false

	switch value[0] {
	case '-':
		negative = true
		value = value[1:]

	case '+':
		value = value[1:]
	}

	if value == "" {
		return 0, errors.New("invalid amount")
	}

	integerPart := value
	fractionPart := ""

	if separator := strings.IndexAny(value, ",."); separator >= 0 {
		if strings.ContainsAny(value[:separator], ",.") || strings.ContainsAny(value[separator+1:], ",.") {
			return 0, errors.New("invalid amount")
		}

		integerPart = value[:separator]
		fractionPart = value[separator+1:]

		if fractionPart == "" {
			return 0, errors.New("invalid amount")
		}
	}

	if integerPart == "" {
		integerPart = "0"
	}

	if !allDigits(integerPart) || (fractionPart != "" && !allDigits(fractionPart)) || len(fractionPart) > 2 {
		return 0, errors.New("invalid amount")
	}

	fractionPart += strings.Repeat("0", 2-len(fractionPart))

	integer, err := strconv.ParseUint(integerPart, 10, 64)
	if err != nil {
		return 0, err
	}

	fraction, err := strconv.ParseUint(fractionPart, 10, 64)
	if err != nil {
		return 0, err
	}

	if integer > (math.MaxUint64-fraction)/100 {
		return 0, errors.New("amount overflow")
	}

	total := integer*100 + fraction

	if negative {
		if total > uint64(math.MaxInt64)+1 {
			return 0, errors.New("amount overflow")
		}

		if total == uint64(math.MaxInt64)+1 {
			return math.MinInt64, nil
		}

		return -int64(total), nil
	}

	if total > math.MaxInt64 {
		return 0, errors.New("amount overflow")
	}

	return int64(total), nil
}

func allDigits(value string) bool {
	if value == "" {
		return false
	}

	for i := 0; i < len(value); i++ {
		if value[i] < '0' || value[i] > '9' {
			return false
		}
	}

	return true
}

func isMCC(value string) bool {
	if len(value) != 4 {
		return false
	}

	for i := 0; i < len(value); i++ {
		if value[i] < '0' || value[i] > '9' {
			return false
		}
	}

	return true
}

func field(record []string, position int) string {
	if position < 0 || position >= len(record) {
		return ""
	}

	return record[position]
}

func recordLine(reader *csv.Reader, record []string) int {
	if len(record) == 0 {
		return 0
	}

	line, _ := reader.FieldPos(0)

	return line
}

func (result *Result) addError(line int, message string) {
	if len(result.Errors) >= maxRowErrors {
		result.ErrorsTruncated = true

		return
	}

	result.Errors = append(result.Errors, RowError{Line: line, Message: message})
}

func stripBOM(input io.Reader) io.Reader {
	reader := bufio.NewReader(input)
	prefix, err := reader.Peek(len(utf8BOM))
	if err == nil && string(prefix) == utf8BOM {
		_, _ = reader.Discard(len(utf8BOM))
	}

	return reader
}

func normalizeSpace(value string) string {
	return strings.Join(strings.Fields(value), " ")
}
