package i18n

import "testing"

func TestEnglishCatalogueDefinesEveryMessage(t *testing.T) {
	t.Parallel()

	for _, key := range AllKeys() {
		if got := English(key); got == "" {
			t.Errorf("English(%q) is empty", key)
		}
	}
}
