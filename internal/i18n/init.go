package i18n

import (
	"github.com/goccy/go-json"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

type Service interface {
	T(lang string, key string, params map[string]any) string
}

type I18nService struct {
	bundle *i18n.Bundle
}

func NewInitI18nService() *I18nService {
	bundle := i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)

	bundle.MustLoadMessageFile("./internal/i18n/en.json")
	bundle.MustLoadMessageFile("./internal/i18n/de.json")

	return &I18nService{bundle: bundle}
}

func (g *I18nService) T(lang string, key string, params map[string]any) string {
	localizer := i18n.NewLocalizer(g.bundle, lang)
	msg, err := localizer.Localize(&i18n.LocalizeConfig{
		MessageID:    key,
		TemplateData: params,
	})

	if err != nil {
		return key
	}

	return msg
}
