package formatter

import (
	"strings"
	"testing"
)

func TestFormatterBasic(t *testing.T) {
	unformatted := `
function test(a:i32,b:i32):i32{
let x:i32=10
if x>5{
print("ok")
}
return a+b
}
`
	formatted := Format(unformatted)
	if !strings.Contains(formatted, "    let x") {
		t.Errorf("expected indented body, got:\n%s", formatted)
	}
	if !strings.Contains(formatted, "        print(\"ok\")") {
		t.Errorf("expected nested block indented with 8 spaces, got:\n%s", formatted)
	}
}
