package migrator

import (
	"reflect"
	"testing"
)

func TestSplitStatements(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		script string
		want   []string
	}{
		{
			name:   "empty",
			script: "  \n\t ",
			want:   nil,
		},
		{
			name:   "single statement without trailing semicolon",
			script: "CREATE TABLE users (id SERIAL)",
			want:   []string{"CREATE TABLE users (id SERIAL)"},
		},
		{
			name: "several statements",
			script: `CREATE TABLE a (id INT);
CREATE INDEX a_id ON a (id);
`,
			want: []string{
				"CREATE TABLE a (id INT)",
				"CREATE INDEX a_id ON a (id)",
			},
		},
		{
			name:   "semicolon inside a string literal",
			script: `INSERT INTO t (v) VALUES ('a;b');`,
			want:   []string{`INSERT INTO t (v) VALUES ('a;b')`},
		},
		{
			name:   "escaped quotes inside a literal",
			script: `INSERT INTO t (v) VALUES ('it''s; fine');`,
			want:   []string{`INSERT INTO t (v) VALUES ('it''s; fine')`},
		},
		{
			name:   "semicolon inside a quoted identifier",
			script: `CREATE TABLE "weird; name" (id INT);`,
			want:   []string{`CREATE TABLE "weird; name" (id INT)`},
		},
		{
			name: "line and block comments are kept but never split",
			script: `-- first; statement
CREATE TABLE a (id INT); /* trailing; comment */
SELECT 1;`,
			want: []string{
				"-- first; statement\nCREATE TABLE a (id INT)",
				"/* trailing; comment */\nSELECT 1",
			},
		},
		{
			name: "dollar quoted function body",
			script: `CREATE FUNCTION f() RETURNS trigger AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
SELECT 1;`,
			want: []string{
				`CREATE FUNCTION f() RETURNS trigger AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql`,
				"SELECT 1",
			},
		},
		{
			name:   "tagged dollar quote",
			script: `DO $body$ SELECT 1; $body$;`,
			want:   []string{`DO $body$ SELECT 1; $body$`},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got, err := SplitStatements(test.script)
			if err != nil {
				t.Fatalf("SplitStatements() error = %v", err)
			}

			if !reflect.DeepEqual(got, test.want) {
				t.Fatalf("SplitStatements() = %#v, want %#v", got, test.want)
			}
		})
	}
}

func TestSplitStatementsErrors(t *testing.T) {
	t.Parallel()

	tests := map[string]string{
		"unterminated string":         `INSERT INTO t VALUES ('abc);`,
		"unterminated block comment":  `SELECT 1 /* comment`,
		"unterminated dollar quoting": `CREATE FUNCTION f() AS $$ SELECT 1;`,
	}

	for name, script := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if _, err := SplitStatements(script); err == nil {
				t.Fatal("SplitStatements() error = nil, want error")
			}
		})
	}
}
