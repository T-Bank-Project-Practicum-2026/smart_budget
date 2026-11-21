const getKey = (key: string, type: "MUTATION" | "QUERY") =>
  `LIST_${key}_${type}`;

export const QueryKeys = {
  me: getKey("me", "QUERY"),
  imports: getKey("imports", "QUERY"),
  importTBank: getKey("importTBank", "MUTATION"),
};
