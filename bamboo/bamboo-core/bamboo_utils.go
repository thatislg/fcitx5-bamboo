/*
 * Bamboo - A Vietnamese Input method editor
 * Copyright (C) Luong Thanh Lam <ltlam93@gmail.com>
 *
 * This software is licensed under the MIT license. For more information,
 * see <https://github.com/BambooEngine/bamboo-core/blob/master/LICENSE>.
 */

// =============================================================================
// CHÚ THÍCH THÊM (Tiếng Việt)
// =============================================================================
// File: bamboo_utils.go
// Mục đích: Chứa toàn bộ logic nền tảng của engine xử lý tiếng Việt. Các hàm
//           trong file này được bamboo.go gọi để tìm kiếm transformation đích,
//           tách từ/âm tiết theo cấu trúc CVC, kiểm tra tính hợp lệ, sinh
//           transformation (undo/fallback/shortcut) và xử lý các trường hợp đặc biệt.
//           Đây là file quan trọng nhất về mặt thuật toán trong bamboo-core.
// =============================================================================

package bamboo

import (
	"regexp"
	"unicode"
)

// findLastAppendingTrans tìm transformation APPENDING cuối cùng trong composition.
//
// Input: Danh sách transformations.
// Output: Transformation APPENDING cuối cùng; nil nếu không có.
// Logic: Duyệt ngược từ cuối composition, tìm transformation có EffectType == Appending.
// Ví dụ: Trong a-as (a + sắc), tìm thấy transformation của 'a'.
func findLastAppendingTrans(composition []*Transformation) *Transformation {
	for i := len(composition) - 1; i >= 0; i-- {
		var trans = composition[i]
		if trans.Rule.EffectType == Appending {
			return trans
		}
	}
	return nil
}

// newAppendingTrans tạo một transformation APPENDING mới.
//
// Input: key — phím gốc; isUpperCase — có phải chữ hoa không.
// Output: Transformation APPENDING với Rule.Key = Rule.EffectOn = Rule.Result = key.
// Logic: Mỗi phím chữ cái thường tạo ra APPENDING transformation, giữ nguyên phím.
func newAppendingTrans(key rune, isUpperCase bool) *Transformation {
	return &Transformation{
		IsUpperCase: isUpperCase,
		Rule: Rule{
			Key:        key,
			EffectOn:   key,
			EffectType: Appending,
			Result:     key,
		},
	}
}

// generateAppendingTrans tìm rule APPENDING khớp với phím, hoặc tạo mới.
//
// Input: rules — danh sách rule khả dụng; lowerKey — phím thường hóa; isUpperCase.
// Output: Transformation APPENDING.
// Logic:
//   - Duyệt rules, nếu có rule APPENDING khớp key → trả về rule đó.
//   - Nếu không có → gọi newAppendingTrans tạo mới.
func generateAppendingTrans(rules []Rule, lowerKey rune, isUpperCase bool) *Transformation {
	for _, rule := range rules {
		if rule.Key == lowerKey && rule.EffectType == Appending {
			var _isUpperCase = isUpperCase || unicode.IsUpper(rule.EffectOn)
			rule.EffectOn = unicode.ToLower(rule.EffectOn)
			rule.Result = rule.EffectOn
			return &Transformation{
				IsUpperCase: _isUpperCase,
				Rule:        rule,
			}
		}
	}
	return newAppendingTrans(lowerKey, isUpperCase)
}

// filterAppendingComposition lọc chỉ lấy các transformation APPENDING.
//
// Input: composition đầy đủ.
// Output: Slice chỉ chứa các transformation có EffectType == Appending.
// Logic: Duyệt composition, chọn các transformation gốc (không phải tone/mark).
func filterAppendingComposition(composition []*Transformation) []*Transformation {
	var appendingTransformations []*Transformation
	for _, trans := range composition {
		if trans.Rule.EffectType == Appending {
			appendingTransformations = append(appendingTransformations, trans)
		}
	}
	return appendingTransformations
}

// findRootTarget tìm transformation gốc nhất trong chuỗi Target.
//
// Input: Một transformation.
// Output: Transformation gốc nhất (không còn Target trỏ tới transformation khác).
// Logic: Đệ quy theo chuỗi Target cho đến khi Target == nil.
// Ý nghĩa: Khi áp dụng nhiều dấu lên cùng một chữ, cần tìm đến gốc để áp dụng rule mới.
func findRootTarget(target *Transformation) *Transformation {
	if target.Target == nil {
		return target
	} else {
		return findRootTarget(target.Target)
	}
}

// isValid kiểm tra xem một âm tiết có phải tiếng Việt hợp lệ không.
//
// Input: composition (thường là một âm tiết); inputIsFullComplete — kiểm tra hoàn chỉnh.
// Output: true nếu âm tiết hợp lệ.
// Logic:
//   - Nếu <= 1 transformation → luôn hợp lệ.
//   - Kiểm tra dấu thanh cuối có hợp lệ với phụ âm cuối không (hasValidTone).
//   - Tách CVC, Flatten từng phần, gọi isValidCVC kiểm tra chính tả.
//
// Ví dụ: "nguyeenx" → false; "nguyen" → true.
func isValid(composition []*Transformation, inputIsFullComplete bool) bool {
	if len(composition) <= 1 {
		return true
	}
	// last tone checking
	for i := len(composition) - 1; i >= 0; i-- {
		if composition[i].Rule.EffectType == ToneTransformation {
			var lastTone = Tone(composition[i].Rule.Effect)
			if !hasValidTone(composition, lastTone) {
				return false
			}
			break
		}
	}
	// spell checking
	var fc, vo, lc = extractCvcTrans(composition)
	var flattenMode = VietnameseMode | LowerCase | ToneLess
	return isValidCVC(Flatten(fc, flattenMode), Flatten(vo, flattenMode), Flatten(lc, flattenMode), inputIsFullComplete)
}

// getRightMostVowels lấy nhóm nguyên âm bên phải nhất trong âm tiết.
//
// Input: composition (thường là một âm tiết).
// Output: Slice các transformation nguyên âm bên phải nhất.
// Logic: Gọi extractCvcTrans để tách C-V-C, trả về phần vowel.
func getRightMostVowels(composition []*Transformation) []*Transformation {
	var _, vo, _ = extractCvcTrans(composition)
	return vo
}

// findToneTarget xác định nguyên âm nào sẽ được đặt dấu thanh.
//
// Input: Một âm tiết (composition); stdStyle — có dùng kiểu chuẩn không.
// Output: Transformation nguyên âm đích.
// Logic phức tạp:
//   - 1 nguyên âm → target là nguyên âm đó.
//   - 2 nguyên âm + stdStyle: ưu tiên ơ/ê; nếu không có, có phụ âm cuối → thứ 2, không → thứ 1.
//   - 2 nguyên âm + không stdStyle: có phụ âm cuối → thứ 2; kiểm tra oa/oe/uy/ue/uo → thứ 2; ngược lại → thứ 1.
//   - 3 nguyên âm (uye...) → target thứ 3 nếu là "uye", ngược lại thứ 2.
//
// Ví dụ:
//   - "hoa" (oa, không phụ âm cuối) → target thứ 2 (a)
//   - "hiệu" (có ê, stdStyle) → target ê
//   - "huyền" (3 nguyên âm, uye) → target e (thứ 3)
func findToneTarget(composition []*Transformation, stdStyle bool) *Transformation {
	if len(composition) == 0 {
		return nil
	}
	var target *Transformation
	var _, vo, lc = extractCvcTrans(composition)
	var vowels = filterAppendingComposition(vo)
	if len(vowels) == 1 {
		target = vowels[0]
	} else if len(vowels) == 2 && stdStyle {
		for _, trans := range vo {
			if trans.Rule.Result == 'ơ' || trans.Rule.Result == 'ê' {
				if trans.Target != nil {
					target = trans.Target
				} else {
					target = trans
				}
			}
		}
		if target == nil {
			if len(lc) > 0 {
				target = vowels[1]
			} else {
				target = vowels[0]
			}
		}
	} else if len(vowels) == 2 {
		if len(lc) > 0 {
			target = vowels[1]
		} else {
			var str = Flatten(vowels, EnglishMode|LowerCase|ToneLess|MarkLess)
			if str == "oa" || str == "oe" || str == "uy" || str == "ue" || str == "uo" {
				target = vowels[1]
			} else {
				target = vowels[0]
			}
		}
	} else if len(vowels) == 3 {
		if Flatten(vowels, EnglishMode|LowerCase|ToneLess|MarkLess) == "uye" {
			target = vowels[2]
		} else {
			target = vowels[1]
		}
	}
	return target
}

// hasValidTone kiểm tra dấu thanh có hợp lệ với phụ âm cuối không.
//
// Input: composition (một âm tiết); tone — dấu thanh cần kiểm tra.
// Output: true nếu dấu thanh hợp lệ.
// Logic:
//   - Sắc (Acute) và nặng (Dot) luôn hợp lệ.
//   - Các phụ âm cuối c, k, p, t, ch không thể đi với huyền/hỏi/ngã.
//
// Ý nghĩa: Đảm bảo không tạo ra âm tiết vô nghĩa như "màc", "tỏp".
func hasValidTone(composition []*Transformation, tone Tone) bool {
	if tone == ToneNone || tone == ToneAcute || tone == ToneDot {
		return true
	}
	var _, _, lc = extractCvcTrans(composition)
	if len(lc) == 0 {
		return true
	}
	var lastConsonants = Flatten(lc, EnglishMode|LowerCase)

	// These consonants have to go with ACUTE, DOT accents
	var dotWithConsonants = []string{"c", "k", "p", "t", "ch"}
	for _, s := range dotWithConsonants {
		if s == lastConsonants {
			return false
		}
	}
	return true
}

// getLastToneTransformation lấy transformation dấu thanh cuối cùng trong âm tiết.
//
// Input: composition.
// Output: Transformation TONE cuối có Target != nil; nil nếu không có.
// Logic: Duyệt ngược composition, tìm transformation có EffectType == ToneTransformation và Target != nil.
func getLastToneTransformation(composition []*Transformation) *Transformation {
	for i := len(composition) - 1; i >= 0; i-- {
		var t = composition[i]
		if t.Rule.EffectType == ToneTransformation && t.Target != nil {
			return t
		}
	}
	return nil
}

// isFree kiểm tra xem một transformation đã bị gắn transformation cùng loại chưa.
//
// Input: composition; trans — transformation cần kiểm tra; effectType — loại effect.
// Output: true nếu target chưa bị gắn transformation cùng loại.
// Logic: Duyệt composition, nếu có transformation nào có Target == trans và EffectType == effectType → false.
// Ý nghĩa: Một nguyên âm chỉ có thể có 1 dấu thanh và 1 dấu phụ.
func isFree(composition []*Transformation, trans *Transformation, effectType EffectType) bool {
	for _, t := range composition {
		if t.Target == trans && t.Rule.EffectType == effectType {
			return false
		}
	}
	return true
}

// extractAtomicTrans tách composition theo loại ký tự (nguyên âm/phụ âm) đệ quy.
//
// Input: composition; last — tích lũy; lastIsVowel — loại ký tự đang tách.
// Output: (phần còn lại, phần đã tách).
// Logic: Duyệt ngược composition, tách cho đến khi gặp ký tự khác loại (nguyên âm vs phụ âm).
// Ý nghĩa: Là bước cơ bản để tách CVC.
func extractAtomicTrans(composition, last []*Transformation, lastIsVowel bool) ([]*Transformation, []*Transformation) {
	if len(composition) == 0 {
		return composition, last
	}
	var tmp = composition[len(composition)-1]
	if tmp != nil && tmp.Target == nil && lastIsVowel != IsVowel(tmp.Rule.Result) {
		return composition, last
	}
	return extractAtomicTrans(composition[:len(composition)-1], append([]*Transformation{composition[len(composition)-1]}, last...), lastIsVowel)
}

/*
   Separate a string into smaller parts: first consonant (or head), vowel,
   last consonant (if any).
*/
// extractCvcAppendingTrans tách appending list thành cấu trúc CVC.
//
// Input: composition chỉ gồm các transformation APPENDING.
// Output: (firstConsonant, vowel, lastConsonant).
// Logic:
//   1. Gọi extractAtomicTrans tách phụ âm cuối và phần còn lại.
//   2. Gọi tiếp extractAtomicTrans trên phần còn lại để tách first consonant và vowel.
//   3. Xử lý đặc biệt gi và qu: Nếu g+i hoặc q+u (không phải gie+phụ âm),
//      chuyển i/u sang first consonant (gi và qu là phụ âm đầu).
// Ví dụ:
//   - ['g', 'i', 'a'] → ['g', 'i'], ['a'], [] (gi = phụ âm đầu)
//   - ['q', 'u', 'a'] → ['q', 'u'], ['a'], [] (qu = phụ âm đầu)
//   - ['g', 'i', 'e', 'n', 'g'] → ['g'], ['i', 'e'], ['n', 'g'] (gie+phụ âm = giữ nguyên)
func extractCvcAppendingTrans(composition []*Transformation) ([]*Transformation, []*Transformation, []*Transformation) {
	head, lastConsonant := extractAtomicTrans(composition, nil, false)
	firstConsonant, vowel := extractAtomicTrans(head, nil, true)
	if len(lastConsonant) > 0 && len(vowel) == 0 && len(firstConsonant) == 0 {
		firstConsonant = lastConsonant
		vowel = nil
		lastConsonant = nil
	}

	// 'gi' and 'qu' are considered qualified consonants.
	// We want something like this:
	//     ['g', 'ia', ''] -> ['gi', 'a', '']
	//     ['q', 'ua', ''] -> ['qu', 'a', '']
	// except:
	//     ['g', 'ie', 'ng'] -> ['g', 'ie', 'ng']
	if len(firstConsonant) == 1 && len(vowel) > 0 && ((firstConsonant[0].Rule.Result == 'g' && vowel[0].Rule.Result == 'i' && len(vowel) > 1 &&
		!(vowel[1].Rule.Result == 'e' && len(lastConsonant) > 0)) ||
		(firstConsonant[0].Rule.Result == 'q' && vowel[0].Rule.Result == 'u')) {
		firstConsonant = append(firstConsonant, vowel[0])
		vowel = vowel[1:]
	}
	return firstConsonant, vowel, lastConsonant
}

// extractCvcTrans tách toàn bộ composition thành cấu trúc CVC (bao gồm cả tone/mark).
//
// Input: Toàn bộ composition.
// Output: (firstConsonant, vowel, lastConsonant) bao gồm cả transformation tác động.
// Logic:
//  1. Tách composition thành appendingList (các APPENDING) và transMap (map target→transformations).
//  2. Gọi extractCvcAppendingTrans trên appendingList.
//  3. Nối các transformation tác động (tone/mark) vào đúng nhóm C/V/L.
//
// Ý nghĩa: Để kiểm tra chính tả, cần biết dấu thanh/mark nằm trong nhóm nào.
func extractCvcTrans(composition []*Transformation) ([]*Transformation, []*Transformation, []*Transformation) {
	var transMap = map[*Transformation][]*Transformation{}
	var appendingList []*Transformation
	for _, trans := range composition {
		if trans.Target == nil {
			appendingList = append(appendingList, trans)
		} else {
			transMap[trans.Target] = append(transMap[trans.Target], trans)
		}
	}
	var fc, vo, lc = extractCvcAppendingTrans(appendingList)
	for _, t := range fc {
		fc = append(fc, transMap[t]...)
	}
	for _, t := range vo {
		vo = append(vo, transMap[t]...)
	}
	for _, t := range lc {
		lc = append(lc, transMap[t]...)
	}
	return fc, vo, lc
}

// extractLastWordWithPunctuationMarks tách từ cuối cùng trong composition kèm dấu câu.
//
// Input: composition; effectKeys — danh sách phím đặc biệt.
// Output: (previous, last) — phần trước từ cuối và từ cuối.
// Logic: Duyệt ngược composition, tìm ký tự space. Nếu gặp space ở cuối → trả về toàn bộ và nil.
//
//	Nếu gặp space ở giữa → tách tại đó.
func extractLastWordWithPunctuationMarks(composition []*Transformation, effectKeys []rune) ([]*Transformation, []*Transformation) {
	for i := len(composition) - 1; i >= 0; i-- {
		var canvas = getCanvas(composition[i:], EnglishMode)
		if len(canvas) == 0 {
			continue
		}
		var c = canvas[0]
		if IsSpace(c) {
			if i == len(composition)-1 {
				return composition, nil
			}
			return composition[:i+1], composition[i+1:]
		}
	}
	return nil, composition
}

// extractLastWord tách từ cuối cùng trong composition.
//
// Input: composition; effectKeys — danh sách phím đặc biệt.
// Output: (previous, last) — phần trước từ cuối và từ cuối.
// Logic: Duyệt ngược composition, tìm ký tự không phải chữ cái và không nằm trong effectKeys
//
//	(thường là space hoặc dấu câu). Tách tại đó.
func extractLastWord(composition []*Transformation, effectKeys []rune) ([]*Transformation, []*Transformation) {
	for i := len(composition) - 1; i >= 0; i-- {
		var canvas = getCanvas(composition[i:], VietnameseMode|LowerCase|ToneLess|MarkLess)
		if len(canvas) == 0 {
			continue
		}
		var c = canvas[0]
		if !IsAlpha(c) && !inKeyList(effectKeys, c) {
			if i == len(composition)-1 {
				return composition, nil
			}
			return composition[:i+1], composition[i+1:]
		}
	}
	return nil, composition
}

// extractLastSyllable tách âm tiết cuối trong composition (đảm bảo hợp lệ).
//
// Input: composition.
// Output: (previous, last) — phần trước âm tiết cuối và âm tiết cuối hợp lệ.
// Logic:
//  1. Gọi extractLastWord để lấy từ cuối.
//  2. Trong từ cuối, tìm vị trí chia tách sao cho đoạn cuối là âm tiết hợp lệ.
//
// Ý nghĩa: Một từ có thể chứa nhiều âm tiết; hàm này tách để chỉ xử lý âm tiết cuối.
func extractLastSyllable(composition []*Transformation) ([]*Transformation, []*Transformation) {
	var previous, last = extractLastWord(composition, nil)
	var anchor = 0
	for i := range last {
		if !isValid(last[anchor:i+1], false) {
			anchor = i
		}
	}
	if anchor > 0 {
		previous = append(previous, last[:anchor]...)
	}
	return previous, last[anchor:]
}

// findMarkTarget tìm transformation đích và rule cho dấu phụ (móc, râu, trăng...).
//
// Input: Một âm tiết; rules — danh sách rule khả dụng.
// Output: (target, rule) — target là transformation bị tác động, rule là rule sẽ áp dụng.
// Logic:
//  1. Duyệt ngược composition.
//  2. Với mỗi transformation, thử áp dụng rule MARK. Kiểm tra:
//     - Result của transformation khớp EffectOn của rule.
//     - Effect > 0 (có dấu phụ).
//     - Áp dụng rule không làm chuỗi giống nhau (tránh lặp).
//     - isValid sau khi thử áp dụng.
//  3. Trả về target và rule đầu tiên hợp lệ.
func findMarkTarget(composition []*Transformation, rules []Rule) (*Transformation, Rule) {
	var str = Flatten(composition, VietnameseMode)
	for i := len(composition) - 1; i >= 0; i-- {
		var trans = composition[i]
		for _, rule := range rules {
			if rule.EffectType != MarkTransformation {
				continue
			}
			if trans.Rule.Result == rule.EffectOn && rule.Effect > 0 {
				var target = findRootTarget(trans)
				if str == Flatten(append(composition, &Transformation{Target: target, Rule: rule}), VietnameseMode) {
					continue
				}
				var tmp = append(composition, &Transformation{Rule: rule, Target: target})
				if isValid(tmp, false) {
					return target, rule
				}
			}
		}
	}
	return nil, Rule{}
}

// findTarget tìm transformation đích cho một phím (dấu thanh hoặc dấu phụ).
//
// Input: composition; applicableRules — rule khả dụng; flags — cấu hình engine.
// Output: (target, rule) — target cho dấu thanh hoặc dấu phụ.
// Logic:
//  1. Ưu tiên tìm TONE target trước (qua findToneTarget).
//     - Nếu EfreeToneMarking bật: tìm nguyên âm thích hợp.
//     - Nếu không: target là nguyên âm APPENDING cuối.
//  2. Kiểm tra nếu áp dụng rule làm chuỗi không đổi (lặp) → bỏ qua.
//  3. Nếu rule TONE không khớp → gọi findMarkTarget.
//
// Ý nghĩa: Entry point chính để quyết định phím dấu sẽ tác động lên chữ nào.
func findTarget(composition []*Transformation, applicableRules []Rule, flags uint) (*Transformation, Rule) {
	var str = Flatten(composition, VietnameseMode)
	// find tone target
	for _, applicableRule := range applicableRules {
		if applicableRule.EffectType != ToneTransformation {
			continue
		}
		var target *Transformation
		if flags&EfreeToneMarking != 0 {
			if hasValidTone(composition, Tone(applicableRule.Effect)) {
				target = findToneTarget(composition, flags&EstdToneStyle != 0)
			}
		} else if lastAppending := findLastAppendingTrans(composition); lastAppending != nil && IsVowel(lastAppending.Rule.EffectOn) {
			target = lastAppending
		}
		if str == Flatten(append(composition, &Transformation{Target: target, Rule: applicableRule}), VietnameseMode) {
			continue
		}
		if Tone(applicableRule.Effect) == ToneNone && isFree(composition, target, ToneTransformation) &&
			FindToneFromChar(target.Rule.Result) == ToneNone {
			target = nil
		}
		return target, applicableRule
	}
	return findMarkTarget(composition, applicableRules)
}

// generateUndoTransformations sinh các transformation để bỏ dấu/mark khi nhấn lại phím dấu.
//
// Input: composition; rules — rule khả dụng; flags — cấu hình engine.
// Output: Các transformation UNDO để bỏ dấu/mark.
// Logic:
//   - Với rule TONE: tìm target, tạo transformation TONE với Effect = 0 (xóa dấu).
//   - Với rule MARK: tìm target có Result khớp EffectOn, tạo transformation MARK với Effect = 0.
//
// Ý nghĩa: Khi nhấn lại phím dấu (ví dụ: aa trong Telex → a thay vì â), engine cần undo dấu trước đó.
func generateUndoTransformations(composition []*Transformation, rules []Rule, flags uint) []*Transformation {
	var transformations []*Transformation
	var str = Flatten(composition, VietnameseMode|ToneLess|LowerCase)
	for _, rule := range rules {
		if rule.EffectType == ToneTransformation {
			var target *Transformation
			if flags&EfreeToneMarking != 0 {
				if hasValidTone(composition, Tone(rule.Effect)) {
					target = findToneTarget(composition, flags&EstdToneStyle != 0)
				}
			} else if lastAppending := findLastAppendingTrans(composition); lastAppending != nil && IsVowel(lastAppending.Rule.EffectOn) {
				target = lastAppending
			}
			if target == nil {
				continue
			}
			var trans = new(Transformation)
			trans.Target = target
			trans.Rule = Rule{
				EffectType: ToneTransformation,
				Effect:     0,
				Key:        0,
			}
			transformations = append(transformations, trans)
		} else if rule.EffectType == MarkTransformation {
			for i := len(composition) - 1; i >= 0; i-- {
				var trans = composition[i]
				if trans.Rule.Result == rule.EffectOn {
					var target = findRootTarget(trans)
					var trans = new(Transformation)
					trans.Target = target
					trans.Rule = Rule{
						Key:        0,
						EffectType: MarkTransformation,
						Effect:     0,
					}
					if str == Flatten(append(composition, trans), VietnameseMode|ToneLess|LowerCase) {
						continue
					}
					transformations = append(transformations, trans)
				}
			}
		}
	}
	return transformations
}

// Regex cho shortcut uwo+ — khớp pattern uơ hoặc ưo theo sau bởi ít nhất 1 chữ cái.
// Dùng trong bamboo.go applyUowShortcut.
var regUOhTail = regexp.MustCompile(`(uơ|ưo)\p{L}+`)

// Regex cho undo ư → u — khớp ưo hoặc ươ trong âm tiết.
// Dùng trong generateTransformations để xử lý ươ + o → uô.
var regUhO = regexp.MustCompile(`(ưo|ươ)`)

/**
* 1 | o + ff     ->  undo + append       -> of
* 2 | o + fs     ->  override            -> ó
* 3 | o + fz     ->  override            -> o
* 4 | o + z      ->  append              -> oz
* 5 | o + f      ->  tone_grave          -> ò
* 6 | w + w      ->  raw                 -> w
* 7 | (u)wo + w  ->  undo + append       -> uow
* ...
**/

// generateTransformations sinh các transformation cho một phím mới — HÀM QUAN TRỌNG NHẤT FILE.
//
// Input: composition hiện tại; applicableRules — rule khả dụng; flags; lowerKey; isUpperCase.
// Output: Danh sách transformation cho phím này (có thể rỗng).
// Logic:
//  1. Double typing undo: Nếu nhấn lại phím effect (ví dụ: aw rồi lại w trong Telex 2),
//     tạo MARK MarkRaw để undo → trả về ngay.
//  2. Tìm target: Gọi findTarget. Nếu tìm thấy:
//     - Tạo transformation với rule và target.
//     - Nếu không phải MARK → trả về ngay.
//     - Nếu là MARK và hợp lệ → trả về. Nếu không hợp lệ, thử tìm target khác (virtual rule cho uow).
//  3. Shortcut ươ/ưo: Nếu âm tiết chứa ưo/ươ và nhấn o, undo ư → u rồi áp dụng rule mới.
//  4. Fallback undo: Nếu không tìm thấy target, thử undo các effect hiện có, sau đó append phím gốc.
//
// Các ví dụ trong comment tiếng Anh phía trên.
func generateTransformations(composition []*Transformation, applicableRules []Rule, flags uint, lowerKey rune, isUpperCase bool) []*Transformation {
	var transformations []*Transformation
	// Double typing an effect key undoes it and its effects, e.g. w + w -> w (Telex 2)
	if len(composition) > 0 {
		var rule = composition[len(composition)-1].Rule
		if rule.EffectType == Appending && rule.Key == lowerKey && rule.Key != rule.Result {
			transformations = append(transformations, &Transformation{
				Rule: Rule{
					EffectType: MarkTransformation,
					Effect:     uint8(MarkRaw),
					Key:        0,
				},
				Target: composition[len(composition)-1],
			})
			return transformations
		}
	}
	// A target may be applied by many different transformations, e.g. o + o + w -> ơ
	if target, applicableRule := findTarget(composition, applicableRules, flags); target != nil {
		transformations = append(transformations, &Transformation{
			Rule:        applicableRule,
			Target:      target,
			IsUpperCase: isUpperCase,
		})
		if applicableRule.EffectType != MarkTransformation {
			return transformations
		}
		var newComp = append(composition, transformations...)
		if isValid(newComp, true) {
			return transformations
		}
		// Implement the uow typing shortcut by creating a virtual
		// Mark_HORN rule that targets 'u' or 'o'.
		if target, virtualRule := findTarget(newComp, applicableRules, flags); target != nil {
			virtualRule.Key = 0
			return append(transformations, &Transformation{virtualRule, target, false})
		}
	} else {
		// Implement ươ/ưo(i/c/ng) + o -> uô
		if regUhO.MatchString(Flatten(composition, VietnameseMode|ToneLess|LowerCase)) {
			var vowels = filterAppendingComposition(getRightMostVowels(composition))
			var trans = &Transformation{
				Target: vowels[0],
				Rule: Rule{
					EffectType: MarkTransformation,
					Key:        0,
					Effect:     uint8(MarkNone),
				},
			}
			if target, applicableRule := findTarget(append(composition, trans), applicableRules, flags); target != nil && target != vowels[0] {
				transformations = append(transformations, trans)
				transformations = append(transformations, &Transformation{
					Rule:        applicableRule,
					Target:      target,
					IsUpperCase: isUpperCase,
				})
				return transformations
			}
		}
		if undoTrans := generateUndoTransformations(composition, applicableRules, flags); len(undoTrans) > 0 {
			// If an effect key can't find its target, it tries to undo its effects, e.g. ươ + w -> uow
			transformations = append(transformations, undoTrans...)
			transformations = append(transformations, newAppendingTrans(lowerKey, isUpperCase))
		}
	}
	return transformations
}

// generateFallbackTransformations sinh fallback APPENDING khi không có rule nào khớp.
//
// Input: composition; applicableRules; lowerKey; isUpperCase.
// Output: Fallback transformations (ít nhất 1 APPENDING).
// Logic:
//  1. Tạo APPENDING transformation cho key (qua generateAppendingTrans).
//  2. Xử lý các AppendedRules (các rule đi kèm, ví dụ: macro ư có thể append thêm dấu).
func generateFallbackTransformations(composition []*Transformation, applicableRules []Rule, lowerKey rune, isUpperCase bool) []*Transformation {
	var transformations []*Transformation
	var trans = generateAppendingTrans(applicableRules, lowerKey, isUpperCase)
	transformations = append(transformations, trans)
	for _, appendedRule := range trans.Rule.AppendedRules {
		var _isUpperCase = isUpperCase || unicode.IsUpper(appendedRule.EffectOn)
		appendedRule.Key = 0 // this is a virtual key
		appendedRule.EffectOn = unicode.ToLower(appendedRule.EffectOn)
		appendedRule.Result = appendedRule.EffectOn
		transformations = append(transformations, &Transformation{
			Rule:        appendedRule,
			IsUpperCase: _isUpperCase,
		})
	}
	return transformations
}

// breakComposition phân rã âm tiết tiếng Việt thành các phím gốc.
//
// Input: Một âm tiết (composition).
// Output: Composition mới chỉ gồm các phím gốc (không có dấu thanh/phụ).
// Logic: Lặp qua composition, bỏ qua các transformation có Key == 0 (virtual), tạo APPENDING mới từ Rule.Key.
// Ý nghĩa: Chuyển âm tiết tiếng Việt về dạng phím gốc (ví dụ: chuyển → chuyeen).
func breakComposition(composition []*Transformation) []*Transformation {
	var result []*Transformation
	for _, trans := range composition {
		if trans.Rule.Key == 0 {
			continue
		}
		result = append(result, newAppendingTrans(trans.Rule.Key, trans.IsUpperCase))
	}
	return result
}

// refreshLastToneTarget di chuyển dấu thanh sang nguyên âm đúng khi âm tiết thay đổi.
//
// Input: Một âm tiết (composition); stdStyle — có dùng kiểu chuẩn không.
// Output: Các transformation để di chuyển dấu thanh (có thể rỗng).
// Logic:
//  1. Lấy nguyên âm bên phải nhất (getRightMostVowels).
//  2. Lấy transformation dấu thanh cuối.
//  3. Tính lại target đúng (findToneTarget).
//  4. Nếu target thay đổi: tạo transformation xóa dấu cũ + tạo transformation đặt dấu mới.
//
// Ví dụ: chuyr → chuỷ (dấu trên y), sau đó chuyrene → chuyển (dấu chuyển sang e).
func refreshLastToneTarget(composition []*Transformation, stdStyle bool) []*Transformation {
	var transformations []*Transformation
	var rightmostVowels = getRightMostVowels(composition)
	var lastToneTrans = getLastToneTransformation(composition)
	if rightmostVowels == nil || lastToneTrans == nil {
		return nil
	}
	var newToneTarget = findToneTarget(composition, stdStyle)
	if lastToneTrans.Target != newToneTarget {
		lastToneTrans.Target = newToneTarget
		transformations = append(transformations, &Transformation{
			Target: lastToneTrans.Target,
			Rule: Rule{
				Key:        0,
				EffectType: ToneTransformation,
				Effect:     uint8(ToneNone),
			},
		})
		var overrideRule = lastToneTrans.Rule
		overrideRule.Key = 0
		transformations = append(transformations, &Transformation{
			Target: newToneTarget,
			Rule:   overrideRule,
		})
	}
	return transformations
}
