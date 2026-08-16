/*
 * Bamboo - A Vietnamese Input method editor
 * Copyright (C) Luong Thanh Lam <ltlam93@gmail.com>
 *
 * This software is licensed under the MIT license. For more information,
 * see <https://github.com/BambooEngine/bamboo-core/blob/master/LICENSE>.
 */

package bamboo

import (
	"unicode"
)

// =============================================================================
// CHÚ THÍCH THÊM (Tiếng Việt)
// =============================================================================
// File: utils.go
// Mục đích: Các hàm tiện ích xử lý ký tự tiếng Việt được sử dụng rộng rãi
//            trong toàn bộ bamboo-core.
//
// Kiến trúc chính:
//   - Vowels: danh sách tất cả nguyên âm (84 rune), nhóm 6 ký tự/nhóm
//   - marksMaps: bảng "họ" ký tự (a/â/ă, e/ê, o/ô/ơ...)
//   - Các hàm AddTone/AddMark: thêm dấu thanh và dấu mũ/móc
//   - Các hàm kiểm tra: IsVowel, IsVietnameseRune, canProcessKey...
// =============================================================================

// Vowels chứa tất cả các nguyên âm tiếng Việt (84 rune).
//
// Cấu trúc: nhóm 6 ký tự cho mỗi nguyên âm gốc (không dấu + 5 thanh).
// Ví dụ:
//
//	Vị trí 0-5:   a à á ả ã ạ     (a + 5 thanh)
//	Vị trí 6-11:  ă ằ ắ ẳ ẵ ặ    (ă + 5 thanh)
//	Vị trí 12-17: â ầ ấ ẩ ẫ ậ    (â + 5 thanh)
//	...
//
// Pattern quan trọng: pos % 6 cho biết thanh điệu.
//
//	0 = không dấu, 1 = huyền, 2 = sắc, 3 = hỏi, 4 = ngã, 5 = nặng
//
// Được sử dụng bởi FindToneFromChar() và AddToneToChar().
var Vowels = []rune("aàáảãạăằắẳẵặâầấẩẫậeèéẻẽẹêềếểễệiìíỉĩịoòóỏõọôồốổỗộơờớởỡợuùúủũụưừứửữựyỳýỷỹỵ")

// PunctuationMarks chứa các ký tự dấu câu và ký tự đặc biệt.
// Được sử dụng bởi IsPunctuationMark() và IsWordBreakSymbol().
var PunctuationMarks = []rune{
	',', ';', ':', '.', '"', '\'', '!', '?', ' ',
	'<', '>', '=', '+', '-', '*', '/', '\\',
	'_', '~', '`', '@', '#', '$', '%', '^', '&', '(', ')', '{', '}', '[', ']',
	'|',
}

// IsSpace kiểm tra xem ký tự có phải là dấu cách không.
func IsSpace(key rune) bool {
	return key == ' '
}

// IsPunctuationMark kiểm tra xem ký tự có phải là dấu câu không.
func IsPunctuationMark(key rune) bool {
	for _, c := range PunctuationMarks {
		if c == key {
			return true
		}
	}
	return false
}

// IsWordBreakSymbol kiểm tra xem ký tự có phải là ký tự ngắt từ không.
// Ký tự ngắt từ bao gồm: dấu câu HOẶC chữ số.
//
// Ví dụ: ' ', ',', '5' → true; 'a', 'á' → false
func IsWordBreakSymbol(key rune) bool {
	return IsPunctuationMark(key) || ('0' <= key && '9' >= key)
}

// IsVowel kiểm tra xem ký tự có phải là nguyên âm tiếng Việt không.
//
// Ví dụ:
//
//	IsVowel('a') → true
//	IsVowel('ấ') → true
//	IsVowel('b') → false
func IsVowel(chr rune) bool {
	isVowel := false
	for _, v := range Vowels {
		if v == chr {
			isVowel = true
		}
	}
	return isVowel
}

// FindVowelPosition tìm vị trí của ký tự trong danh sách Vowels.
//
// Output: vị trí (0-83) hoặc -1 nếu không phải nguyên âm.
//
// Được sử dụng bởi FindToneFromChar() và AddToneToChar().
func FindVowelPosition(chr rune) int {
	for pos, v := range Vowels {
		if v == chr {
			return pos
		}
	}
	return -1
}

// marksMaps định nghĩa các "họ" ký tự trong tiếng Việt.
//
// Mỗi entry ánh xạ một ký tự gốc → chuỗi chứa các biến thể của họ đó.
// Placeholder '_' đánh dấu vị trí không có ký tự.
//
// Ví dụ:
//
//	'a': "aâă__"  → họ của 'a' gồm: a, â, ă
//	'o': "oô_ơ_"  → họ của 'o' gồm: o, ô, ơ
//	'd': "d___đ"  → họ của 'd' gồm: d, đ
//
// Vị trí trong chuỗi tương ứng với giá trị Mark:
//
//	0 = MarkNone, 1 = MarkHat, 2 = MarkBreve, 3 = MarkHorn, 4 = MarkDash
var marksMaps = map[rune]string{
	'a': "aâă__",
	'â': "aâă__",
	'ă': "aâă__",
	'e': "eê___",
	'ê': "eê___",
	'o': "oô_ơ_",
	'ô': "oô_ơ_",
	'ơ': "oô_ơ_",
	'u': "u__ư_",
	'ư': "u__ư_",
	'd': "d___đ",
	'đ': "d___đ",
}

// getMarkFamily trả về các ký tự trong cùng họ với chr.
//
// Ví dụ:
//
//	getMarkFamily('a') → ['a', 'â', 'ă']
//	getMarkFamily('o') → ['o', 'ô', 'ơ']
//	getMarkFamily('d') → ['d', 'đ']
func getMarkFamily(chr rune) []rune {
	var result []rune
	if s, found := marksMaps[chr]; found {
		for _, c := range s {
			if c != '_' {
				result = append(result, c)
			}
		}
	}
	return result
}

// FindMarkPosition tìm vị trí của chr trong họ ký tự của nó.
//
// Output: vị trí (0-4) hoặc -1 nếu không tìm thấy.
// Vị trí này tương ứng với giá trị Mark.
func FindMarkPosition(chr rune) int {
	if str, found := marksMaps[chr]; found {
		for pos, v := range []rune(str) {
			if v == chr {
				return pos
			}
		}
	}
	return -1
}

// FindMarkFromChar tìm loại Mark của một ký tự.
//
// Input: ký tự cần kiểm tra (ví dụ: 'â', 'ă', 'đ')
// Output: (Mark, true) nếu tìm thấy; (0, false) nếu không.
//
// Ví dụ:
//
//	FindMarkFromChar('â') → (MarkHat, true)    // vị trí 1 trong "aâă__"
//	FindMarkFromChar('ă') → (MarkBreve, true)  // vị trí 2 trong "aâă__"
//	FindMarkFromChar('a') → (MarkNone, true)   // vị trí 0
func FindMarkFromChar(chr rune) (Mark, bool) {
	var pos = FindMarkPosition(chr)
	if pos >= 0 {
		return Mark(pos), true
	}
	return 0, false
}

// AddMarkToTonelessChar thêm dấu mũ/móc vào ký tự không dấu thanh.
//
// Input:
//   - chr: ký tự gốc (không dấu thanh, ví dụ: 'a', 'e', 'o')
//   - mark: loại mark cần thêm (MarkHat, MarkBreve, MarkHorn...)
//
// Output: ký tự sau khi thêm mark, hoặc giữ nguyên nếu không hợp lệ.
//
// Ví dụ:
//
//	AddMarkToTonelessChar('a', 1) → 'â'  // MarkHat
//	AddMarkToTonelessChar('a', 2) → 'ă'  // MarkBreve
//	AddMarkToTonelessChar('a', 3) → 'a' // MarkHorn không hợp lệ cho 'a'
func AddMarkToTonelessChar(chr rune, mark uint8) rune {
	if str, found := marksMaps[chr]; found {
		marks := []rune(str)
		if marks[mark] != '_' {
			return marks[mark]
		}
	}
	return chr
}

// AddMarkToChar thêm dấu mũ/móc vào ký tự, giữ nguyên dấu thanh hiện tại.
//
// Logic:
//  1. Lưu tone hiện tại: tone = FindToneFromChar(chr)
//  2. Bỏ tone: chr = AddToneToChar(chr, 0) → về không dấu
//  3. Thêm mark: chr = AddMarkToTonelessChar(chr, mark)
//  4. Khôi phục tone: chr = AddToneToChar(chr, tone)
//
// Ví dụ:
//
//		AddMarkToChar('á', MarkHat) → 'ấ'
//	  // 'á' (a sắc) → 'a' (bỏ sắc) → 'â' (thêm mũ) → 'ấ' (thêm sắc lại)
func AddMarkToChar(chr rune, mark uint8) rune {
	tone := FindToneFromChar(chr)
	chr = AddToneToChar(chr, 0)
	chr = AddMarkToTonelessChar(chr, mark)
	return AddToneToChar(chr, uint8(tone))
}

// IsAlpha kiểm tra xem ký tự có phải là chữ cái Latin (a-z, A-Z) không.
func IsAlpha(c rune) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

// inKeyList kiểm tra xem key có nằm trong danh sách keys không.
func inKeyList(keys []rune, key rune) bool {
	for _, k := range keys {
		if k == key {
			return true
		}
	}
	return false
}

// FindToneFromChar tìm thanh điệu của một ký tự nguyên âm.
//
// Logic: Tone(pos % 6), trong đó pos là vị trí trong Vowels.
//
// Ví dụ:
//
//	FindToneFromChar('a') → ToneNone   // vị trí 0, 0 % 6 = 0
//	FindToneFromChar('á') → ToneAcute  // vị trí 2, 2 % 6 = 2
//	FindToneFromChar('ạ') → ToneDot    // vị trí 5, 5 % 6 = 5
func FindToneFromChar(chr rune) Tone {
	pos := FindVowelPosition(chr)
	if pos == -1 {
		return ToneNone
	}
	return Tone(pos % 6)
}

// AddToneToChar thêm hoặc đổi dấu thanh cho một ký tự nguyên âm.
//
// Logic:
//  1. Tìm vị trí pos của chr trong Vowels
//  2. Tính currentTone = pos % 6
//  3. Tính offset = tone - currentTone
//  4. Trả về Vowels[pos + offset]
//
// Ví dụ:
//
//	AddToneToChar('a', 2) → 'á'  // a + sắc
//	AddToneToChar('â', 3) → 'ẩ'  // â + hỏi
//	AddToneToChar('ấ', 0) → 'â'  // ấ → bỏ sắc (về â)
func AddToneToChar(chr rune, tone uint8) rune {
	pos := FindVowelPosition(chr)
	if pos > -1 {
		currentTone := pos % 6
		offset := int(tone) - currentTone
		return Vowels[pos+offset]
	} else {
		return chr
	}
}

// canProcessKey kiểm tra xem một phím có cần được engine xử lý không.
//
// Logic:
//  1. Nếu là chữ cái alpha HOẶC trong danh sách effectKeys → true
//  2. Nếu là ký tự ngắt từ (space, số, dấu câu) → false
//  3. Nếu là ký tự tiếng Việt (có dấu) → true
//
// Ví dụ:
//
//	canProcessKey('a', ['s','f']) → true  // alpha
//	canProcessKey('s', ['s','f']) → true  // trong effectKeys
//	canProcessKey(' ', ['s','f']) → false // space ngắt từ
//	canProcessKey('á', ['s','f']) → true  // tiếng Việt
func canProcessKey(lowerKey rune, effectKeys []rune) bool {
	if IsAlpha(lowerKey) || inKeyList(effectKeys, lowerKey) {
		return true
	}
	if IsWordBreakSymbol(lowerKey) {
		return false
	}
	return IsVietnameseRune(lowerKey)
}

// IsVietnameseRune kiểm tra xem ký tự có phải là ký tự tiếng Việt không.
//
// Logic:
//  1. Nếu có dấu thanh (FindToneFromChar != ToneNone) → true
//  2. Hoặc nếu có dấu mũ/móc (chr != AddMarkToTonelessChar(chr, 0)) → true
//
// Ví dụ:
//
//	IsVietnameseRune('á') → true  // có dấu sắc
//	IsVietnameseRune('â') → true  // có dấu mũ
//	IsVietnameseRune('a') → false // không có gì đặc biệt
//	IsVietnameseRune('b') → false
func IsVietnameseRune(lowerKey rune) bool {
	// lowerKey = unicode.ToLower(lowerKey)
	if FindToneFromChar(lowerKey) != ToneNone {
		return true
	}
	return lowerKey != AddMarkToTonelessChar(lowerKey, 0)
}

// HasAnyVietnameseRune kiểm tra xem chuỗi có chứa ký tự tiếng Việt không.
func HasAnyVietnameseRune(word string) bool {
	for _, chr := range word {
		if IsVietnameseRune(unicode.ToLower(chr)) {
			return true
		}
	}
	return false
}

// HasAnyVietnameseVower kiểm tra xem chuỗi có chứa nguyên âm tiếng Việt không.
func HasAnyVietnameseVower(word string) bool {
	for _, chr := range word {
		if IsVowel(unicode.ToLower(chr)) {
			return true
		}
	}
	return false
}
