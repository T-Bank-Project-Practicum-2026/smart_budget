// @title API «Умный бюджет»
// @version 1.0.0
// @description Авторизация через HttpOnly-cookie. Для POST, PUT и DELETE нужны разрешённый Origin и X-CSRF-Protection: 1.
// @servers.url /
// @externalDocs.description Спецификация OpenAPI 3.1
// @externalDocs.url https://spec.openapis.org/oas/v3.1.0

// @securityDefinitions.apikey accessCookie
// @in cookie
// @name budget_access

// @securityDefinitions.apikey refreshCookie
// @in cookie
// @name budget_refresh
package server
