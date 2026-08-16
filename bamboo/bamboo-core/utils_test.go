package bamboo

import (
	"testing"
)

// =============================================================================
// CHÚ THÍCH THÊM (Tiếng Việt)
// =============================================================================
// File: utils_test.go
// Mục đích: Test suite cho các hàm tiện ích xử lý ký tự tiếng Việt:
//           IsVowel, FindToneFromChar, AddToneToChar, AddMarkToChar.
// =============================================================================

// TestIsVowel kiểm tra hàm IsVowel() nhận diện đúng nguyên âm tiếng Việt.
//
// Cases: 'a', 'á' → true; 'b' → false; duyệt 84 nguyên âm có dấu → tất cả true.
func TestIsVowel(t *testing.T) {
	if IsVowel('a') == false {
		t.Errorf("a is a vowel, but result is false")
	}
	if IsVowel('á') == false {
		t.Errorf("á is a vowel, but result is false")
	}
	if IsVowel('b') {
		t.Errorf("b is not a vowel, but result is true")
	}
	tvowels := string("aàáảãạăằắẳẵặâầấẩẫậeèéẻẽẹêềếểễệiìíỉĩịoòóỏõọôồốổỗộơờớởỡợuùúủũụưừứửữựyỳýỷỹỵ")
	for _, v := range tvowels {
		if IsVowel(v) == false {
			t.Errorf("%c is a vowel, but the result is false", v)
		}
	}
}

// TestGetToneFromChar kiểm tra hàm FindToneFromChar() trích xuất đúng dấu thanh.
//
// Cases: e → ToneNone, è → ToneGrave, é → ToneAcute, ẽ → ToneTilde,
//
//	ẻ → ToneHook, ạ → ToneDot.
func TestGetToneFromChar(t *testing.T) {
	none := FindToneFromChar('e')
	if none != ToneNone {
		t.Errorf("Test none tune. Got %d, expected %d", none, ToneNone)
	}
	grave := FindToneFromChar('è')
	if grave != ToneGrave {
		t.Errorf("Test grave tune. Got %d, expected %d", grave, ToneGrave)
	}
	acute := FindToneFromChar('é')
	if acute != ToneAcute {
		t.Errorf("Test acute tune. Got %d, expected %d", acute, ToneAcute)
	}
	tilde := FindToneFromChar('ẽ')
	if tilde != ToneTilde {
		t.Errorf("Test acute tune. Got %d, expected %d", tilde, ToneTilde)
	}
	hook := FindToneFromChar('ẻ')
	if hook != ToneHook {
		t.Errorf("Test hook tune. Got %d, expected %d", hook, ToneHook)
	}
	dot := FindToneFromChar('ạ')
	if dot != ToneDot {
		t.Errorf("Test dot tune. Got %d, expected %d", dot, ToneDot)
	}
}

// TestAddToneToChar kiểm tra hàm AddToneToChar() thêm dấu thanh đúng.
//
// Cases: a + ToneDot → ạ; y + ToneNone → y (không đổi);
//
//	y + MarkNone → y (không đổi, test thêm AddMarkToChar).
func TestAddToneToChar(t *testing.T) {
	cẠ := AddToneToChar('a', uint8(ToneDot))
	if cẠ != 'ạ' {
		t.Errorf("Test add a dot to char. Got %c, expected %c", cẠ, 'ạ')
	}
	cY := AddToneToChar('y', 0)
	if cY != 'y' {
		t.Errorf("Add TONE_NONE to char y, got %c expected y", cY)
	}
	cY = AddMarkToChar('y', 0)
	if cY != 'y' {
		t.Errorf("Add MARK_NONE to char y, got %c expected y", cY)
	}
}

// TestAddMarkToChar kiểm tra hàm AddMarkToChar() thêm dấu phụ đúng.
//
// Case: ạ + MarkBreve → ặ (thêm dấu trăng vào ạ).
func TestAddMarkToChar(t *testing.T) {
	cẶ := AddMarkToChar('ạ', uint8(MarkBreve))
	if cẶ != 'ặ' {
		t.Errorf("Test add a breve to char. Got %c, expected %c", cẶ, 'ặ')
	}
}
