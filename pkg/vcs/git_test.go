// Copyright 2017 syzkaller project authors. All rights reserved.
// Use of this source code is governed by Apache 2 LICENSE that can be found in the LICENSE file.

package vcs

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestGitParseCommit(t *testing.T) {
	tests := map[string]*Commit{
		`2075b16e32c26e4031b9fd3cbe26c54676a8fcb5
rbtree: include rcu.h
foobar@foobar.de
Fri May 11 16:02:14 2018 -0700
Since commit c1adf20052d8 ("Introduce rb_replace_node_rcu()")
rbtree_augmented.h uses RCU related data structures but does not include
the header file.  It works as long as it gets somehow included before
that and fails otherwise.

Link: http://lkml.kernel.org/r/20180504103159.19938-1-bigeasy@linutronix.de
Signed-off-by: Foo Bad Baz <another@email.de>
Reviewed-by: <yetanother@email.org>
Cc: Unrelated Guy <somewhere@email.com>
Acked-by: Subsystem reviewer <Subsystem@reviewer.com>
Reported-and-tested-by: and@me.com
Reported-and-Tested-by: Name-name <name@name.com>
Tested-by: Must be correct <mustbe@correct.com>
Signed-off-by: Linux Master <linux@linux-foundation.org>
`: {
			Hash:   "2075b16e32c26e4031b9fd3cbe26c54676a8fcb5",
			Title:  "rbtree: include rcu.h",
			Author: "foobar@foobar.de",
			CC: []string{
				"and@me.com",
				"foobar@foobar.de",
				"mustbe@correct.com",
				"name@name.com",
				"subsystem@reviewer.com",
				"yetanother@email.org",
			},
			Date: time.Date(2018, 5, 11, 16, 02, 14, 0, time.FixedZone("", -7*60*60)),
		},
	}
	for input, com := range tests {
		res, err := gitParseCommit([]byte(input))
		if err != nil && com != nil {
			t.Fatalf("want %+v, got error: %v", com, err)
		}
		if err == nil && com == nil {
			t.Fatalf("want error, got commit %+v", res)
		}
		if com == nil {
			continue
		}
		if com.Hash != res.Hash {
			t.Fatalf("want hash %q, got %q", com.Hash, res.Hash)
		}
		if com.Title != res.Title {
			t.Fatalf("want title %q, got %q", com.Title, res.Title)
		}
		if com.Author != res.Author {
			t.Fatalf("want author %q, got %q", com.Author, res.Author)
		}
		if !reflect.DeepEqual(com.CC, res.CC) {
			t.Fatalf("want CC %q, got %q", com.CC, res.CC)
		}
		if !com.Date.Equal(res.Date) {
			t.Fatalf("want date %v, got %v", com.Date, res.Date)
		}
	}
}

func TestGitParseReleaseTags(t *testing.T) {
	input := `
v3.1
v2.6.12
v2.6.39
v3.0
v3.10
v2.6.13
v3.11
v3.19
v3.9
v3.2
v4.9
v2.6.32
v4.0
voo
v1.foo
v10.2.foo
v1.2.
v1.
`
	want := []string{
		"v4.9",
		"v4.0",
		"v3.19",
		"v3.11",
		"v3.10",
		"v3.9",
		"v3.2",
		"v3.1",
		"v3.0",
		"v2.6.39",
		"v2.6.32",
		"v2.6.13",
		"v2.6.12",
	}
	got, err := gitParseReleaseTags([]byte(input))
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got bad tags\ngot:  %+v\nwant: %+v", got, want)
	}
}

func TestGitExtractFixTags(t *testing.T) {
	commits, err := gitExtractFixTags(strings.NewReader(extractFixTagsInput), extractFixTagsEmail)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(extractFixTagsOutput, commits); diff != "" {
		t.Fatal(diff)
	}
}

const extractFixTagsEmail = "\"syzbot\" <syzbot@my.mail.com>"

var extractFixTagsOutput = []FixCommit{
	{"8e4090902540da8c6e8f", "dashboard/app: bump max repros per bug to 10"},
	{"8e4090902540da8c6e8f", "executor: remove dead code"},
	{"a640a0fc325c29c3efcb", "executor: remove dead code"},
	{"8e4090902540da8c6e8fa640a0fc325c29c3efcb", "pkg/csource: fix string escaping bug"},
	{"6dd701dc797b23b8c761", "When freeing a lockf struct that already is part of a linked list, make sure to"},
}

var extractFixTagsInput = `
commit 73aba437a774237b1130837b856f3b40b3ec3bf0 (HEAD -> master, origin/master)
Author: me <foo@bar.com>
Date:   Fri Dec 22 19:59:56 2017 +0100

    dashboard/app: bump max repros per bug to 10
    
    Reported-by: syzbot+8e4090902540da8c6e8f@my.mail.com

commit 26cd53f078db858a6ccca338e13e7f4d1d291c22
Author: me <foo@bar.com>
Date:   Fri Dec 22 13:42:27 2017 +0100

    executor: remove dead code
    
    Reported-by: syzbot+8e4090902540da8c6e8f@my.mail.com
    Reported-by: syzbot <syzbot+a640a0fc325c29c3efcb@my.mail.com>

commit 7b62abdb0abadbaf7b3f3a23ab4d78485fbf9059
Author: Dmitry Vyukov <dvyukov@google.com>
Date:   Fri Dec 22 11:59:09 2017 +0100

    pkg/csource: fix string escaping bug
    
    Reported-and-tested-by: syzbot+8e4090902540da8c6e8fa640a0fc325c29c3efcb@my.mail.com

commit 47546510aa98d3fbff3291a5dc3cefe712e70394
Author: anton <openbsd@openbsd.org>
Date:   Sat Oct 6 21:12:23 2018 +0000

    When freeing a lockf struct that already is part of a linked list, make sure to
    update the next pointer for the preceding lock. Prevents a double free panic.
    
    ok millert@
    Reported-by: syzbot+6dd701dc797b23b8c761@my.mail.com
`
