package utils

import (
	"github.com/nicksnyder/go-i18n/i18n"
	"go-web-boilerplate/model"
	"go-web-boilerplate/mlog"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
)

var T i18n.TranslateFunc
var TDefault i18n.TranslateFunc
var locales map[string]string = make(map[string]string)
var settings model.LocalizationSettings


func TranslationsPerInit() error {
	T = TfuncWithFallback("zh-CN")
	TDefault = TfuncWithFallback("en")

	if err := InitTranslationsWithDir("i18n"); err != nil {
		return err
	}

	return nil
}


func InitTranslations(localizationSettings model.LocalizationSettings) error {
	settings = localizationSettings

	var err error
	T, err = GetTranslationsBySystemLocale()
	return err
}


func GetTranslationsBySystemLocale() (i18n.TranslateFunc, error) {
	locale := *settings.DefaultServerLocale
	if _, ok := locales[locale]; !ok {
		mlog.Error(fmt.Sprintf("Failed to load system translations for '%v' attempting to fall back to '%v'", locale, model.DEFAULT_LOCALE))
		locale = model.DEFAULT_LOCALE
	}

	if locales[locale] == "" {
		return nil, fmt.Errorf("Failed to load system translations for '%v'", model.DEFAULT_LOCALE)
	}

	translations := TfuncWithFallback(locale)
	if translations == nil {
		return nil, fmt.Errorf("Failed to load system translations")
	}

	mlog.Info(fmt.Sprintf("Loaded system translations for '%v' from '%v'", locale, locales[locale]))
	return translations, nil
}

func InitTranslationsWithDir(dir string) error {
	i18nDirectory, found := FindDir(dir)

	if !found {
		return fmt.Errorf("unable to find i18n directory")
	}

	files, _ := ioutil.ReadDir(i18nDirectory)

	for _ ,f := range files {
		if filepath.Ext(f.Name()) == ".json" {
			filename := f.Name()
			locales[strings.Split(filename, ".")[0]] = filepath.Join(i18nDirectory, filename)

			if err := i18n.LoadTranslationFile(filepath.Join(i18nDirectory, filename)); err != nil {
				return err
			}
		}
	}

	return nil
}

func TfuncWithFallback(pref string) i18n.TranslateFunc {
	t, _ := i18n.Tfunc(pref)
	return func(translationID string, args ...interface{}) string {
		if translated := t(translationID, args...); translated != translationID {
			return translated
		}

		t, _ := i18n.Tfunc(model.DEFAULT_LOCALE)
		return t(translationID, args...)
	}
}