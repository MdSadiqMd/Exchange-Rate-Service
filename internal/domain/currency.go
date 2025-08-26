package domain

type Currency struct {
	Code   string `json:"code"`
	Name   string `json:"name"`
	Symbol string `json:"symbol"`
}

var SupportedCurrencies = []Currency{
	{"USD", "US Dollar", "$"},
	{"INR", "Indian Rupee", "₹"},
	{"EUR", "Euro", "€"},
	{"JPY", "Japanese Yen", "¥"},
	{"GBP", "British Pound", "£"},
}

func IsValidCurrency(code string) bool {
	for _, currency := range SupportedCurrencies {
		if currency.Code == code {
			return true
		}
	}
	return false
}
