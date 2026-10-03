package pr_test

import (
	"testing"

	"github.com/gitwise/backend/internal/service/pr"
)

func TestParseUnifiedDiff(t *testing.T) {
	rawDiff := `diff --git a/pkg/auth/token.go b/pkg/auth/token.go
index 1111111..2222222 100644
--- a/pkg/auth/token.go
+++ b/pkg/auth/token.go
@@ -10,6 +10,8 @@ package auth
-func VerifyToken(token string) bool {
+func VerifyToken(token string, secret string) bool {
+    // Added secret parameter
     return true
 }
diff --git a/pkg/auth/token_test.go b/pkg/auth/token_test.go
new file mode 100644
index 0000000..3333333
--- /dev/null
+++ b/pkg/auth/token_test.go
@@ -0,0 +1,10 @@
+package auth_test
+
+import "testing"
+
+func TestVerifyToken(t *testing.T) {}
`

	files := pr.ParseUnifiedDiff(rawDiff)
	if len(files) != 2 {
		t.Fatalf("expected 2 parsed files, got %d", len(files))
	}

	f1 := files[0]
	if f1.Path != "pkg/auth/token.go" {
		t.Errorf("expected path pkg/auth/token.go, got %s", f1.Path)
	}
	if f1.Status != "modified" {
		t.Errorf("expected status modified, got %s", f1.Status)
	}
	if f1.Additions < 1 || f1.Deletions < 1 {
		t.Errorf("expected additions and deletions, got +%d -%d", f1.Additions, f1.Deletions)
	}

	f2 := files[1]
	if f2.Path != "pkg/auth/token_test.go" {
		t.Errorf("expected path pkg/auth/token_test.go, got %s", f2.Path)
	}
	if f2.Status != "added" {
		t.Errorf("expected status added, got %s", f2.Status)
	}
}
