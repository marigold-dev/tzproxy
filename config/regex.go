package config

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/rs/zerolog/log"
)

func parseRegexRoutes(values []string) map[string][]*regexp.Regexp {
	allMethods := []string{
		http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete,
		http.MethodPatch, http.MethodHead, http.MethodOptions,
	}

	regexRoutes := make(map[string][]*regexp.Regexp)
	for _, route := range values {
		if !containsPrefix(route, allMethods) {
			for _, method := range allMethods {
				regex, err := regexp.Compile(route)
				if err != nil {
					log.Fatal().Err(err).Msg("unable to compile regex")
				}

				regexRoutes[method] = append(regexRoutes[method], regex)
			}
			continue
		}

		for _, method := range allMethods {
			if strings.HasPrefix(strings.ToUpper(route), method) {
				url := strings.TrimPrefix(route, method)
				regex, err := regexp.Compile(url)
				if err != nil {
					log.Fatal().Err(err).Msg("unable to compile regex")
				}

				regexRoutes[method] = append(regexRoutes[method], regex)
				break
			}
		}
	}

	return regexRoutes
}

func containsPrefix(s string, prefixes []string) bool {
	for _, prefix := range prefixes {
		if strings.HasPrefix(s, prefix) {
			return true
		}
	}

	return false
}
