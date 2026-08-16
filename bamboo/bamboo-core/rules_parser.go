/*
 * Bamboo - A Vietnamese Input method editor
 * Copyright (C) Luong Thanh Lam <ltlam93@gmail.com>
 *
 * This software is licensed under the MIT license. For more information,
 * see <https://github.com/BambooEngine/bamboo-core/blob/master/LICENSE>.
 */

package bamboo

import (
	"regexp"
	"strings"
)

// =============================================================================
// CHÚ THÍCH THÊM (Tiếng Việt)
// =============================================================================
// File: rules_parser.go
// Mục đích: Parse định nghĩa kiểu gõ từ input_method_def.go thành các Rule
//            struct để BambooEngine sử dụng.
//
// Luồng xử lý chính:
//   InputMethodDefinitions (map) --ParseRules()--> []Rule
//
// Ví dụ:
//   "s": "DauSac"        → 1 Rule (ToneTransformation)
//   "w": "UOA_ƯƠĂ"       → 18 Rules (MarkTransformation × 3 ký tự × 6 tone)
//   "]": "__ư"           → 1 Rule (Appending)
//
// Kiến trúc:
//   - Tone: enum 6 thanh điệu (None, Grave, Acute, Hook, Tilde, Dot)
//   - Mark: enum 5 dấu mũ/móc (Hat, Breve, Horn, Dash, Raw)
//   - EffectType: enum 4 loại hiệu ứng
//   - Rule: struct chứa thông tin một rule đơn lẻ
//   - InputMethod: struct chứa toàn bộ rules của một kiểu gõ
// =============================================================================

// tones ánh xạ tên hành động (từ input_method_def.go) sang giá trị Tone.
// Được sử dụng trong ParseRules() để phân loại rule có thanh điệu.
var tones = map[string]Tone{
	"XoaDauThanh": ToneNone,  // Xóa dấu thanh (phím 'z')
	"DauSac":      ToneAcute, // Thêm dấu sắc (phím 's')
	"DauHuyen":    ToneGrave, // Thêm dấu huyền (phím 'f')
	"DauNga":      ToneTilde, // Thêm dấu ngã (phím 'x')
	"DauNang":     ToneDot,   // Thêm dấu nặng (phím 'j')
	"DauHoi":      ToneHook,  // Thêm dấu hỏi (phím 'r')
}

// EffectType định nghĩa loại hiệu ứng của một rule.
type EffectType int

const (
	// Appending: Thêm ký tự vào chuỗi, không biến đổi.
	// Ví dụ: Telex 2 dùng ']' để thêm 'ư'.
	Appending EffectType = iota << 0

	// MarkTransformation: Biến đổi dấu mũ/móc của nguyên âm.
	// Ví dụ: 'aa' → 'â', 'aw' → 'ă'.
	MarkTransformation EffectType = iota

	// ToneTransformation: Thêm/xóa dấu thanh điệu.
	// Ví dụ: 'as' → 'á', 'az' → 'a' (xóa dấu).
	ToneTransformation EffectType = iota

	// Replacing: Thay thế ký tự (ít dùng).
	Replacing EffectType = iota
)

// Mark định nghĩa các dấu mũ/móc có thể thêm vào nguyên âm.
// Ví dụ: Hat (â, ê, ô), Horn (ư, ơ), Breve (ă).
type Mark uint8

const (
	MarkNone  Mark = iota // Không có dấu mũ/móc
	MarkHat   Mark = iota // Dấu mũ: â, ê, ô
	MarkBreve Mark = iota // Dấu móc ngược: ă
	MarkHorn  Mark = iota // Dấu móc: ư, ơ
	MarkDash  Mark = iota // Dấu gạch ngang: đ
	MarkRaw   Mark = iota // Không xử lý đặc biệt
)

// Tone định nghĩa 6 thanh điệu trong tiếng Việt.
type Tone uint8

const (
	ToneNone  Tone = iota // Không dấu (a)
	ToneGrave Tone = iota // Dấu huyền (à)
	ToneAcute Tone = iota // Dấu sắc (á)
	ToneHook  Tone = iota // Dấu hỏi (ả)
	ToneTilde Tone = iota // Dấu ngã (ã)
	ToneDot   Tone = iota // Dấu nặng (ạ)
)

// Rule định nghĩa một rule đơn lẻ trong kiểu gõ.
//
// Ví dụ: khi người dùng bấm 's' sau 'a', rule sẽ biến 'a' thành 'á'.
//
// Các trường:
//   - Key: Phím kích hoạt (ví dụ: 's', 'w')
//   - Effect: Giá trị Tone hoặc Mark (tùy EffectType)
//   - EffectType: Loại hiệu ứng (ToneTransformation, MarkTransformation...)
//   - EffectOn: Ký tự bị tác động (ví dụ: 'a' trong "as")
//   - Result: Ký tự kết quả (ví dụ: 'á')
//   - AppendedRules: Rules phụ cho Telex 2/W (thêm nhiều ký tự)
type Rule struct {
	Key           rune
	Effect        uint8 // (Tone, Mark)
	EffectType    EffectType
	EffectOn      rune
	Result        rune
	AppendedRules []Rule
}

func (r *Rule) SetTone(tone Tone) {
	r.Effect = uint8(tone)
}

func (r *Rule) SetMark(mark Mark) {
	r.Effect = uint8(mark)
}

func (r *Rule) GetTone() Tone {
	return Tone(r.Effect)
}

func (r *Rule) GetMark() Mark {
	return Mark(r.Effect)
}

// InputMethod chứa toàn bộ rules của một kiểu gõ (ví dụ: Telex).
// Được tạo ra bởi ParseInputMethod() và sử dụng bởi BambooEngine.
//
// Các trường:
//   - Name: Tên kiểu gõ ("Telex", "Telex 2", "Telex W")
//   - Rules: Toàn bộ rules đã parse
//   - SuperKeys: Các phím đặc biệt (ví dụ: 'w' cho ư/ơ)
//   - ToneKeys: Các phím thêm dấu thanh (s, f, r, x, j)
//   - AppendingKeys: Các phím append (cho Telex 2/W)
//   - Keys: Tất cả các phím trong kiểu gõ
type InputMethod struct {
	Name          string
	Rules         []Rule
	SuperKeys     []rune
	ToneKeys      []rune
	AppendingKeys []rune
	Keys          []rune
}

// ParseInputMethod là entry point để parse một kiểu gõ từ định nghĩa.
//
// Input:
//   - imDef: map chứa tất cả định nghĩa kiểu gò (từ GetInputMethodDefinitions())
//   - imName: tên kiểu gò cần parse (ví dụ: "Telex")
//
// Output:
//   - InputMethod: struct chứa toàn bộ rules của kiểu gò đó
//
// Ví dụ:
//
//	imDef := GetInputMethodDefinitions()
//	telexIM := ParseInputMethod(imDef, "Telex")
func ParseInputMethod(imDef map[string]InputMethodDefinition, imName string) InputMethod {
	var inputMethods = parseInputMethods(imDef)
	if inputMethod, found := inputMethods[imName]; found {
		return inputMethod
	}
	return InputMethod{}
}

// parseInputMethods parse tất cả các kiểu gò trong imDef.
//
// Input: imDef — map chứa định nghĩa tất cả kiểu gò
// Output: map[string]InputMethod — ánh xạ tên kiểu gò → InputMethod đã parse
//
// Logic:
//  1. Lặp qua từng kiểu gò trong imDef
//  2. Với mỗi entry (phím → hành động), gọi ParseRules() để tạo rules
//  3. Nếu hành động chứa "uo", đánh dấu phím là SuperKey
//  4. Phân loại keys thành ToneKeys và AppendingKeys
//  5. Trả về map kết quả
func parseInputMethods(imDef map[string]InputMethodDefinition) map[string]InputMethod {
	var inputMethods = make(map[string]InputMethod, len(imDef))
	for name, imDefinition := range imDef {
		var im InputMethod
		im.Name = name
		for keyStr, line := range imDefinition {
			var keys = []rune(keyStr)
			if len(keys) == 0 {
				continue
			}
			var key = keys[0]
			im.Rules = append(im.Rules, ParseRules(key, line)...)
			if strings.Contains(strings.ToLower(line), "uo") {
				im.SuperKeys = append(im.SuperKeys, key)
			}
			im.Keys = append(im.Keys, key)
		}
		for _, rule := range im.Rules {
			if rule.EffectType == Appending {
				im.AppendingKeys = append(im.AppendingKeys, rule.Key)
			}
			if rule.EffectType == ToneTransformation {
				im.ToneKeys = append(im.ToneKeys, rule.Key)
			}
		}
		inputMethods[name] = im
	}
	return inputMethods
}

// ParseRules phân loại và parse một hành động thành các Rule.
//
// Input:
//   - key: phím bấm (ví dụ: 's', 'w')
//   - line: chuỗi hành động từ input_method_def.go (ví dụ: "DauSac", "UOA_ƯƠĂ")
//
// Output: []Rule — danh sách các rule tương ứng
//
// Logic phân loại:
//  1. Nếu line match tones map (ví dụ: "DauSac") → ToneTransformation
//  2. Nếu không match → gọi ParseTonelessRules() để parse mark/append
//
// Ví dụ:
//
//	ParseRules('s', "DauSac") → [Rule{Key:'s', EffectType:ToneTransformation, Effect:ToneAcute}]
//	ParseRules('w', "UOA_ƯƠĂ") → 18 Rule (MarkTransformation)
func ParseRules(key rune, line string) []Rule {
	var rules []Rule
	if tone, ok := tones[line]; ok {
		var rule Rule
		rule.Key = key
		rule.EffectType = ToneTransformation
		rule.Effect = uint8(tone)
		rules = append(rules, rule)
	} else {
		rules = ParseTonelessRules(key, line)
	}
	return rules
}

// regDsl là regex để parse chuỗi mark transformation.
//
// Pattern: `([a-zA-Z]+)_(\p{L}+)([_\p{L}]*)`
//   - Group 1: Các chữ cái Latin không dấu (ví dụ: "UOA")
//   - Group 2: Dấu gạch dưới + chữ Unicode (ví dụ: "_ƯƠĂ")
//   - Group 3: Phần còn lại (optional, cho Telex 2/W)
//
// Ví dụ:
//
//	"UOA_ƯƠĂ"       → Group 1="UOA", Group 2="ƯƠĂ", Group 3=""
//	"UOA_ƯƠĂ__Ư"    → Group 1="UOA", Group 2="ƯƠĂ", Group 3="__Ư"
var regDsl = regexp.MustCompile(`([a-zA-Z]+)_(\p{L}+)([_\p{L}]*)`)

// ParseTonelessRules parse các rule không phải tone (mark hoặc append).
//
// Input:
//   - key: phím bấm
//   - line: chuỗi hành động (ví dụ: "UOA_ƯƠĂ", "__ư")
//
// Output: []Rule
//
// Logic:
//  1. Thử match regDsl → parse mark transformation
//  2. Nếu không match → thử parse appending rule
func ParseTonelessRules(key rune, line string) []Rule {
	var rules []Rule
	if regDsl.MatchString(line) {
		parts := regDsl.FindStringSubmatch(strings.ToLower(line))
		effectiveOns := []rune(parts[1])
		results := []rune(parts[2])
		for i, effectiveOn := range effectiveOns {
			effect, found := FindMarkFromChar(results[i])
			if !found {
				continue
			}
			rules = append(rules, ParseToneLessRule(key, effectiveOn, results[i], effect)...)
		}
		if rule, ok := getAppendingRule(key, parts[3]); ok {
			rules = append(rules, rule)
		}

	} else if rule, ok := getAppendingRule(key, line); ok {
		rules = append(rules, rule)
	}
	return rules
}

// ParseToneLessRule tạo các rule chi tiết cho mark transformation.
//
// Đây là hàm phức tạp nhất trong file. Nó tạo ra nhiều rule cho từng
// trường hợp thanh điệu, vì engine cần biết cách biến đổi cả khi
// nguyên âm đã có dấu thanh.
//
// Input:
//   - key: phím bấm (ví dụ: 'w')
//   - effectiveOn: ký tự gốc bị tác động (ví dụ: 'u')
//   - result: ký tự kết quả (ví dụ: 'ư')
//   - effect: loại mark (MarkHorn, MarkBreve...)
//
// Output: []Rule — nhiều rule cho từng trường hợp
//
// Logic:
//  1. Lấy "họ" ký tự (getMarkFamily) — các biến thể của effectiveOn
//  2. Nếu ký tự = result → tạo 1 rule đơn giản
//  3. Nếu ký tự là nguyên âm → tạo 6 rules (1 cho mỗi tone)
//  4. Nếu ký tự khác → tạo 1 rule mark transformation
//
// Tại sao tạo 6 rules cho nguyên âm?
//
//	Vì khi người dùng gõ "uw" → "ư", có thể sau đó thêm dấu → "ứ".
//	Engine cần biết rule nào áp dụng cho từng trạng thái dấu thanh.
func ParseToneLessRule(key, effectiveOn, result rune, effect Mark) []Rule {
	var rules []Rule
	var tones = []Tone{ToneNone, ToneDot, ToneAcute, ToneGrave, ToneHook, ToneTilde}
	for _, chr := range getMarkFamily(effectiveOn) {
		if chr == result {
			var rule Rule
			rule.Key = key
			rule.EffectType = MarkTransformation
			rule.Effect = 0
			rule.EffectOn = result
			rule.Result = effectiveOn
			rules = append(rules, rule)
		} else if IsVowel(chr) {
			for tone := range tones {
				var rule Rule
				rule.Key = key
				rule.EffectType = MarkTransformation
				rule.EffectOn = AddToneToChar(chr, uint8(tone))
				rule.Effect = uint8(effect)
				rule.Result = AddToneToChar(result, uint8(tone))
				rules = append(rules, rule)
			}
		} else {
			var rule Rule
			rule.Key = key
			rule.EffectType = MarkTransformation
			rule.EffectOn = chr
			rule.Effect = uint8(effect)
			rule.Result = result
			rules = append(rules, rule)
		}
	}
	return rules
}

// regDslAppending là regex để parse appending rules (Telex 2/W).
//
// Pattern: `(_?)_(\p{L}+)`
//   - Group 1: Có thể là "_" hoặc rỗng (phân biệt hoa/thường)
//   - Group 2: Chữ Unicode cần append
//
// Ví dụ:
//
//	"__ư" → Group 1="_", Group 2="ư" (append chữ thường)
//	"_Ư"  → Group 1="", Group 2="Ư" (append chữ hoa)
var regDslAppending = regexp.MustCompile(`(_?)_(\p{L}+)`)

// getAppendingRule tạo một rule để append ký tự (không biến đổi).
//
// Input:
//   - key: phím bấm
//   - value: chuỗi appending (ví dụ: "__ư", "_Ư")
//
// Output:
//   - Rule: rule với EffectType = Appending
//   - bool: true nếu parse thành công
//
// Được sử dụng cho Telex 2/W khi người dùng muốn thêm ký tự
// đặc biệt bằng phím ngoặc vuông.
func getAppendingRule(key rune, value string) (Rule, bool) {
	var rule Rule
	if regDslAppending.MatchString(value) {
		parts := regDslAppending.FindStringSubmatch(value)
		chars := []rune(parts[2])
		rule.Key = key
		rule.EffectType = Appending
		rule.EffectOn = chars[0]
		rule.Result = chars[0]
		if len(chars) > 1 {
			for _, chr := range chars[1:] {
				rule.AppendedRules = append(rule.AppendedRules, Rule{
					Key:        key,
					EffectType: Appending,
					EffectOn:   chr,
					Result:     chr,
				})
			}
		}
		return rule, true
	}
	return rule, false
}
