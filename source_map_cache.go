package symbolic

/*
#include <stdlib.h>
#include <string.h>
#include "include/symbolic.h"
*/
import "C"
import (
	"runtime"
	"unsafe"
)

type SourceMapCache struct {
	ssmc *C.SymbolicSourceMapCache
}

func NewSourceMapCache(source, sourceMap string) (*SourceMapCache, error) {
	cs := C.CString(source)
	defer C.free(unsafe.Pointer(cs))
	csm := C.CString(sourceMap)
	defer C.free(unsafe.Pointer(csm))

	C.symbolic_err_clear()
	ssmc := C.symbolic_sourcemapcache_from_bytes(cs, C.strlen(cs), csm, C.strlen(csm))
	err := checkErr()

	if err != nil {
		return nil, err
	}

	s := &SourceMapCache{
		ssmc: ssmc,
	}

	runtime.SetFinalizer(s, free)

	return s, nil
}

func (s *SourceMapCache) Lookup(line, col, contextLines uint32) (*SourceMapCacheToken, error) {
	C.symbolic_err_clear()
	match := C.symbolic_sourcemapcache_lookup_token(s.ssmc, C.uint32_t(line), C.uint32_t(col), C.uint32_t(contextLines))
	err := checkErr()

	if err != nil {
		return nil, err
	}

	defer C.symbolic_sourcemapcache_token_match_free(match)

	smct := newSourceMapCacheToken(match)

	return smct, nil
}

func free(s *SourceMapCache) {
	C.symbolic_sourcemapcache_free(s.ssmc)
}

type SourceMapCacheToken struct {
	Line         int
	Col          int
	Src          string
	Name         string
	FunctionName string
	ContextLine  string
	PreContext   []string
	PostContext  []string
}

func newSourceMapCacheToken(match *C.SymbolicSmTokenMatch) *SourceMapCacheToken {
	pre := make([]string, match.pre_context.len)
	for i, s := range unsafe.Slice(match.pre_context.strs, match.pre_context.len) {
		pre[i] = decodeStr(&s)
	}

	post := make([]string, match.post_context.len)
	for i, s := range unsafe.Slice(match.post_context.strs, match.post_context.len) {
		post[i] = decodeStr(&s)
	}

	return &SourceMapCacheToken{
		Line:         int(match.line),
		Col:          int(match.col),
		Src:          decodeStr(&match.src),
		Name:         decodeStr(&match.name),
		FunctionName: decodeStr(&match.function_name),
		ContextLine:  decodeStr(&match.context_line),
		PreContext:   pre,
		PostContext:  post,
	}
}
