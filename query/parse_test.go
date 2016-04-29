// Copyright 2016 Google Inc. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package query

import (
	"reflect"
	"regexp/syntax"
	"testing"
)

func mustParseRE(s string) *syntax.Regexp {
	r, err := syntax.Parse(s, 0)
	if err != nil {
		panic(err)
	}

	return r
}

func TestParseQuery(t *testing.T) {
	type testcase struct {
		in     string
		out    Q
		hasErr bool
	}

	for _, c := range []testcase{
		{"sub-pixel", &Substring{Pattern: "sub-pixel"}, false},
		{"abc", &Substring{Pattern: "abc"}, false},
		{"\"abc bcd\"", &Substring{Pattern: "abc bcd"}, false},
		{"abc bcd", &And{[]Q{
			&Substring{Pattern: "abc"},
			&Substring{Pattern: "bcd"},
		}}, false},
		{"-abc", &Not{&Substring{Pattern: "abc"}}, false},
		{"regex:a.b", nil, true},

		{"abccase:yes", &Substring{Pattern: "abccase:yes"}, false},
		{"file:abc", &Substring{Pattern: "abc", FileName: true}, false},
		{"branch:pqr", &Branch{Name: "pqr"}, false},

		{"file:helpers\\.go byte", &And{[]Q{
			&Substring{Pattern: "helpers.go", FileName: true},
			&Substring{Pattern: "byte"},
		}}, false},
		{"(abc def)", &And{[]Q{
			&Substring{Pattern: "abc"},
			&Substring{Pattern: "def"},
		}}, false},
		{"(abc def", nil, true},
		{"regex:abc[p-q]", &Regexp{Regexp: mustParseRE("abc[p-q]")}, false},
		{"repo:go", &Repo{"go"}, false},

		{"abc.*def", &Regexp{Regexp: mustParseRE("abc.*def")}, false},
		{"abc\\.\\*def", &Substring{Pattern: "abc.*def"}, false},
		{"(abc)", &Regexp{Regexp: mustParseRE("(abc)")}, false},

		// case
		{"abc case:yes", &Substring{Pattern: "abc", CaseSensitive: true}, false},
		{"abc case:auto", &Substring{Pattern: "abc", CaseSensitive: false}, false},
		{"ABC case:auto", &Substring{Pattern: "ABC", CaseSensitive: true}, false},
		{"ABC case:\"auto\"", &Substring{Pattern: "ABC", CaseSensitive: true}, false},
		// errors.
		{"\"abc", nil, true},
		{"\"a\\", nil, true},
		{"case:foo", nil, true},
		{"", nil, true},
	} {
		q, err := Parse(c.in)
		if c.hasErr != (err != nil) {
			t.Errorf("Parse(%s): error %v, value %v", c.in, err, q)
		} else if q != nil {
			if !reflect.DeepEqual(q, c.out) {
				t.Errorf("Parse(%s): got %v want %v", c.in, q, c.out)
			}
		}
	}
}

func TestTokenize(t *testing.T) {
	type testcase struct {
		in   string
		typ  int
		text string
	}

	cases := []testcase{
		{"file:bla", tokFile, "bla"},
		{"file:bla ", tokFile, "bla"},
		{"(abc def) ", tokParenOpen, "("},
		{"(abcdef)", tokText, "(abcdef)"},
		{"(ab(c)def) ", tokText, "(ab(c)def)"},
		{"(ab\\ def) ", tokText, "(ab\\ def)"},
		{") ", tokParenClose, ")"},
		{"abc) ", tokText, "abc"},
		{"file:\"bla\"", tokFile, "bla"},
		{"\"file:bla\"", tokText, "file:bla"},
		{"\\", tokError, ""},
	}
	for _, c := range cases {
		tok, err := nextToken([]byte(c.in))
		if err != nil {
			tok = &token{Type: tokError}
		}
		if tok.Type != c.typ {
			t.Errorf("%s: got type %d, want %d", c.in, tok.Type, c.typ)
			continue
		}

		if string(tok.Text) != c.text {
			t.Errorf("%s: got text %q, want %q", c.in, tok.Text, c.text)
		}
	}
}
