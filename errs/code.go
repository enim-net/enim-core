package errs

import (
	"fmt"
	"regexp"
	"sync"
	"sync/atomic"
)

const (
	coreMax = 99
	appMin  = 100
	maxNum  = 999
)

type Code struct {
	cat Category
	num int
	key string
}

var (
	prefix   atomic.Value
	prefixRe = regexp.MustCompile(`^[A-Z][A-Z0-9]{1,7}$`)
	regMu    sync.Mutex
	registry = map[string]string{} // "CC-NNN" -> key, to catch duplicates
)

func init() { prefix.Store("APP") }

func SetPrefix(p string) {
	if !prefixRe.MatchString(p) {
		panic(fmt.Sprintf("errs: invalid prefix %q (want 2-8 uppercase letters/digits)", p))
	}
	prefix.Store(p)
}

func Prefix() string {
	return prefix.Load().(string)
}

func Define(cat Category, num int, key string) Code {
	if num < appMin || num > maxNum {
		panic(fmt.Sprintf("errs: app code number must be %d-%d, got %d (000-099 are reserved for go-core)", appMin, maxNum, num))
	}
	return define(cat, num, key)
}

func defineCore(cat Category, num int, key string) Code {
	if num > coreMax {
		panic(fmt.Sprintf("errs: core code number must be 000-%03d, got %d", coreMax, num))
	}
	return define(cat, num, key)
}

func define(cat Category, num int, key string) Code {
	c := Code{cat: cat, num: num, key: key}
	id := c.ID()
	regMu.Lock()
	defer regMu.Unlock()
	if existing, ok := registry[id]; ok {
		panic(fmt.Sprintf("errs: duplicate code %s (%s and %s)", id, existing, key))
	}
	registry[id] = key
	return c
}

// ID is the code without prefix, e.g. "04-101".
func (c Code) ID() string { return fmt.Sprintf("%s-%03d", c.cat.Code, c.num) }

// String is the full code, e.g. "ISP-04-101".
func (c Code) String() string { return Prefix() + "-" + c.ID() }

// Error makes Code usable as an error value.
func (c Code) Error() string { return c.String() }

func (c Code) Category() Category { return c.cat }
func (c Code) Key() string        { return c.key }
func (c Code) HTTPStatus() int    { return c.cat.HTTPStatus }

// New creates an *Error with this code and message params.
func (c Code) New(params map[string]any) *Error { return New(c, params) }

// Wrap creates an *Error with this code that keeps cause for logging.
func (c Code) Wrap(cause error, params map[string]any) *Error { return Wrap(cause, c, params) }
