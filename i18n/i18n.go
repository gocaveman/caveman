// Translation tooling for internationalized sites/pages/components.
//
// Groups and Keys
//
// Groups are logical units of translation, loosely corresponding to an application or Go package.
// Group names can be any Go string but it is recommended that they are alphanumeric and specifically
// should not including whitespace, a colon or other punctuation characters.
//
// Keys are more loose in their requirments and can either be a token or the original language text.
// For example, you can use "title_this_is_a_test" as a key or "This is a Test".  The former is
// more specific and can be used to avoid confusion between the same text appearing in different
// contexts and thus needing different translations; whereas the latter is easier to work with when
// rapidly putting in content since you can avoid providing any translation and the English text
// will still appear correctly and a corresopnding translation file can be created later.  Pick your poison,
// but it's recommended that the larger your project, the more you should consider using surrogate tokens
// for translation keys (i.e. "title_something_here" not "Something Here").
//
// Locales
//
// While no format is strictly enforced for locale names, RFC 3066 and ISO 639 should be used.
// Locale names are treated as case-insensitive (implementations should simply convert to lower case)
// but no other change is performed on locale strings, they are otherwise compared using
// normal string Go equality comparison.  Apply the following rules to avoid confusion:
//
// Always use the shortest possible locale string representation which accurately represents the locale.
// ISO 639.1 has two-letter language codes and ISO 639.2 has three-letter codes - always
// prefer the two-letter code where it exists.  E.g. English is "en" not "eng", Finnish is
// "fi" not "fin", however Filipino is "fil" as there is no ISO 639.1 (two-letter) code for it.
//
// A "subtag" should be used (with a dash) where it is necessary to distinguish between language variations.
// E.g. if you are providing a translation in Spanish you should simply use "es" as the locale.
// However if a Castilian Spanish version is needed for Spain, you can use "es-es" (Spanish - Spain),
// "en-gb" would be used for (English - United Kingdom), and so on.  ISO 3166 two-letter country
// codes should be used (following the "shortest possible representation" concept.)
//
package i18n

import (
	"log"

	"github.com/gocaveman/caveman/webutil"
)

// documentation points:
// - needs to support vastly different sizes of scale from a few dozen or hundred words to large databases of text
// - registration and override mechanism so translations can be plugged in
// - organizing translations into groups so different packages can each provide translations in an appropriate namespace
//   and avoid collisions.
// - the option to use regular English or other default language text as the key or a surrogate identifier, e.g. you can
//   use either "This is a test." or "title_this_is_a_test" as the key.  Each has its pros and cons - using default text
//   is faster to prototype with because you don't need to maintain a translations file just to see the first language;
//   whereas using a surrogate unique id/token ensures translations are specific and individual labels for example don't
//   lose their context.  (Using "Name" as a key could be very confusing for example because it may be the same in English
//   but different in other languages depending on the context.)
// - string replacements (give example)
// - multiple or custom storage format, files or database or other arbitrary
// - default behavior for how pages can easily be translated, but also customizable by replacing out just the logic of
//   what decides the locale for a page (put name of struct or method here)
// - support for determining what locales a page is translated into
// - editor for visual translations ui, with pluggable storage mechanisms (i.e can write to file or db)
// - a way to debug and see where text is being pulled from (although it is quite verbose and only appropriate
//   during debugging)

// flat file -> sqlite would be an excellent choice here (although it is actually simpler and
// complex queries are not required, so... we'll see - possibly in the case of large datasets,
// but then the build time would be bad.  yeah, simple in memory cache with LRU or something
// if it gets too big - at least the possibility of plugging that in - probably more the way to go)

// probably want to provide something in the context that can know what the current page's
// locale is, with defaults, and a way to override

// maybe a registry mechanism so other things can provide translations

// at least think through what happens if they want to have one page per translation
// instead of using string lookups, we should facilitate that (although doing it
// with the metadata maybe tricky - but give that some thought too);dont' need to die
// over it but it needs to be feasible if desired.

//////////// AHHHHHHHH - need to have good support for string replacement/arguments - Go templating is probably
// a good choice here, but context needs to be figured out.

// groups also need more thought out - in cases where we do {{$t.T "Stores"}} is this supposed to have an implied group?
// check it check them all?  should it be the "default" group?  Or should we just ignore groups altogether and treat it
// as a flat keyspace

var ErrNotFound = webutil.ErrNotFound

type LocaleGroupTranslator interface {
	T(s string) string // Returns the text translated into one of the target locales or returns back the same text provided as-is.
}

// LocaleTranslator is aware of the current list of locales for the given situation (HTTP request or other)
// and can translate strings into the appropriate text.
type LocaleTranslator interface {
	T2(g, s string) string // Returns the text translated into one of the target locales or returns back the same text provided as-is.
}

type DefaultLocaleTranslator struct {
	Locales    []string
	Translator Translator
}

func (t *DefaultLocaleTranslator) T(g, s string) string {
	ret, err := t.Translator.Translate(g, s, t.Locales...)
	if err != nil {
		log.Printf("Error calling Translator.Translate(%q): %v", s, err)
	}
	return ret
}
