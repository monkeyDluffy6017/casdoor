// Copyright 2021 The Casdoor Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package object

import (
	"log"
)

func (application *Application) GetProviderByCategory(category string) (*Provider, error) {
	providers, err := GetProviders(application.Organization)
	if err != nil {
		return nil, err
	}

	m := map[string]*Provider{}
	for _, provider := range providers {
		if provider.Category != category {
			continue
		}

		m[provider.Name] = provider
	}

	for _, providerItem := range application.Providers {
		if provider, ok := m[providerItem.Name]; ok {
			return provider, nil
		}
	}

	return nil, nil
}

func isProviderItemCountryCodeMatched(providerItem *ProviderItem, countryCode string) bool {
	if len(providerItem.CountryCodes) == 0 {
		return true
	}

	for _, countryCode2 := range providerItem.CountryCodes {
		if countryCode2 == "" || countryCode2 == "All" || countryCode2 == "all" || countryCode2 == countryCode {
			return true
		}
	}
	return false
}

func (application *Application) GetProviderByCategoryAndRule(category string, method string, countryCode string) (*Provider, error) {
	log.Printf("=== GetProviderByCategoryAndRule Debug ===")
	log.Printf("Searching for: Category=%s, Method=%s, CountryCode=%s", category, method, countryCode)
	log.Printf("Organization: %s", application.Organization)

	providers, err := GetProviders(application.Organization)
	if err != nil {
		log.Printf("ERROR getting providers: %v", err)
		return nil, err
	}
	log.Printf("Found %d total providers in organization", len(providers))

	m := map[string]*Provider{}
	for _, provider := range providers {
		log.Printf("Checking provider: Name=%s, Category=%s", provider.Name, provider.Category)
		if provider.Category != category {
			log.Printf("  -> Skipping: category mismatch")
			continue
		}
		log.Printf("  -> Adding to map: category matches")
		m[provider.Name] = provider
	}
	log.Printf("Providers map for category '%s' has %d entries", category, len(m))

	for i, providerItem := range application.Providers {
		log.Printf("Checking application provider[%d]: Name=%s, Rule=%s, CountryCodes=%v", i, providerItem.Name, providerItem.Rule, providerItem.CountryCodes)

		if providerItem.Provider != nil && providerItem.Provider.Category == "SMS" {
			log.Printf("  -> SMS provider, checking country code match...")
			if !isProviderItemCountryCodeMatched(providerItem, countryCode) {
				log.Printf("  -> Country code mismatch, skipping")
				continue
			}
			log.Printf("  -> Country code matched")
		}

		ruleMatched := providerItem.Rule == method || providerItem.Rule == "" || providerItem.Rule == "All" || providerItem.Rule == "all" || providerItem.Rule == "None"
		log.Printf("  -> Rule check: providerItem.Rule='%s', method='%s', ruleMatched=%v", providerItem.Rule, method, ruleMatched)

		if ruleMatched {
			if provider, ok := m[providerItem.Name]; ok {
				log.Printf("  -> MATCH FOUND! Returning provider: %s", provider.Name)
				log.Printf("=== End GetProviderByCategoryAndRule Debug ===")
				return provider, nil
			} else {
				log.Printf("  -> Rule matched but provider not found in map")
			}
		}
	}

	log.Printf("No matching provider found")
	log.Printf("=== End GetProviderByCategoryAndRule Debug ===")
	return nil, nil
}

func (application *Application) GetEmailProvider(method string) (*Provider, error) {
	return application.GetProviderByCategoryAndRule("Email", method, "All")
}

func (application *Application) GetSmsProvider(method string, countryCode string) (*Provider, error) {
	return application.GetProviderByCategoryAndRule("SMS", method, countryCode)
}

func (application *Application) GetStorageProvider() (*Provider, error) {
	return application.GetProviderByCategory("Storage")
}

func (application *Application) getSignupItem(itemName string) *SignupItem {
	for _, signupItem := range application.SignupItems {
		if signupItem.Name == itemName {
			return signupItem
		}
	}
	return nil
}

func (application *Application) IsSignupItemVisible(itemName string) bool {
	signupItem := application.getSignupItem(itemName)
	if signupItem == nil {
		return false
	}

	return signupItem.Visible
}

func (application *Application) IsSignupItemRequired(itemName string) bool {
	signupItem := application.getSignupItem(itemName)
	if signupItem == nil {
		return false
	}

	return signupItem.Required
}

func (si *SignupItem) isSignupItemPrompted() bool {
	return si.Visible && si.Prompted
}

func (application *Application) GetSignupItemRule(itemName string) string {
	signupItem := application.getSignupItem(itemName)
	if signupItem == nil {
		return ""
	}

	return signupItem.Rule
}

func (application *Application) getAllPromptedProviderItems() []*ProviderItem {
	res := []*ProviderItem{}
	for _, providerItem := range application.Providers {
		if providerItem.isProviderPrompted() {
			res = append(res, providerItem)
		}
	}
	return res
}

func (application *Application) getAllPromptedSignupItems() []*SignupItem {
	res := []*SignupItem{}
	for _, signupItem := range application.SignupItems {
		if signupItem.isSignupItemPrompted() {
			res = append(res, signupItem)
		}
	}
	return res
}

func (application *Application) isAffiliationPrompted() bool {
	signupItem := application.getSignupItem("Affiliation")
	if signupItem == nil {
		return false
	}

	return signupItem.Prompted
}

func (application *Application) HasPromptPage() bool {
	providerItems := application.getAllPromptedProviderItems()
	if len(providerItems) != 0 {
		return true
	}

	signupItems := application.getAllPromptedSignupItems()
	if len(signupItems) != 0 {
		return true
	}

	return application.isAffiliationPrompted()
}
