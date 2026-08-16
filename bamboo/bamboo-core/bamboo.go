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
// File: bamboo.go
// Mục đích: Engine chính xử lý bàn phím tiếng Việt. Quản lý trạng thái gõ
//           (composition), xử lý từng phím người dùng nhấn, biến đổi chuỗi
//           phím thô thành chuỗi tiếng Việt có dấu.
//           Đây là file cốt lõi được C++ gọi qua CGO (qua bamboo_utils.go).
// =============================================================================

// Package bamboo implements text processing for Vietnamese
package bamboo

import (
	"unicode"
)

// ---------------------------------------------------------------------------
// Mode — Các chế độ hoạt động của engine (dùng khi lấy output)
// ---------------------------------------------------------------------------
// Mode quyết định GetProcessedString / Flatten sẽ trả về chuỗi dạng gì.
// Ví dụ: VietnameseMode → đầy đủ dấu; ToneLess|MarkLess → không dấu.
// ---------------------------------------------------------------------------
type Mode uint

const (
	VietnameseMode  Mode = 1 << iota // Chế độ tiếng Việt — output có đầy đủ dấu thanh, dấu phụ
	EnglishMode                      // Chế độ tiếng Anh — phím không qua xử lý tiếng Việt
	ToneLess                         // Khi Flatten: bỏ dấu thanh (sắc, huyền, hỏi, ngã, nặng)
	MarkLess                         // Khi Flatten: bỏ dấu phụ (móc, râu, trăng, dấu gạch)
	LowerCase                        // Khi Flatten: ép về chữ thường
	FullText                         // Khi GetProcessedString: lấy toàn bộ composition thay vì chỉ âm tiết cuối
	PunctuationMode                  // Xử lý dấu câu đặc biệt (dấu chấm, phẩy...)
	InReverseOrder                   // Thêm phím vào đầu composition (dùng cho kiểu gõ đảo ngược)
)

// ---------------------------------------------------------------------------
// Flags — Các cờ cấu hình engine (bật/tắt tính năng)
// ---------------------------------------------------------------------------
// Flags khác Mode: Flags là cấu hình engine (truyền vào lúc khởi tạo),
// còn Mode là yêu cầu định dạng output.
// ---------------------------------------------------------------------------
const (
	EfreeToneMarking    uint                                                     = 1 << iota // Cho phép đặt dấu thanh tự do (không cần đúng quy tắc vị trí)
	EstdToneStyle                                                                            // Kiểu dấu thanh chuẩn (ví dụ: ư, ơ, v.v.)
	EautoCorrectEnabled                                                                      // Bật tự động sửa chính tả
	EstdFlags           = EfreeToneMarking | EstdToneStyle | EautoCorrectEnabled             // Cấu hình mặc định
)

// ---------------------------------------------------------------------------
// Transformation — Đơn vị biến đổi cơ bản
// ---------------------------------------------------------------------------
// Mỗi phím người dùng nhấn tạo ra một hoặc nhiều Transformation.
// Ví dụ: nhấn 'a' → APPENDING transformation;
// nhấn 's' sau 'a' → TONE transformation với Target trỏ về transformation của 'a'.
// ---------------------------------------------------------------------------
type Transformation struct {
	Rule        Rule            // Quy tắc áp dụng (APPENDING, MARK, TONE...)
	Target      *Transformation // Con trỏ tới transformation bị tác động (nếu là MARK/TONE)
	IsUpperCase bool            // Phím gốc có phải chữ hoa không
}

// ---------------------------------------------------------------------------
// IEngine — Giao diện engine để C++ gọi qua CGO
// ---------------------------------------------------------------------------
// Tất cả các hàm trong interface này đều có bản export trong bamboo_utils.go
// (qua //export) để phía C++ (src/mint/) có thể gọi được.
// ---------------------------------------------------------------------------
type IEngine interface {
	SetFlag(uint)
	GetInputMethod() InputMethod
	ProcessKey(rune, Mode)
	ProcessString(string, Mode)
	GetProcessedString(Mode) string
	IsValid(bool) bool
	CanProcessKey(rune) bool
	RemoveLastChar(bool)
	RestoreLastWord(bool)
	Reset()
}

// ---------------------------------------------------------------------------
// BambooEngine — Triển khai engine xử lý tiếng Việt
// ---------------------------------------------------------------------------
// composition: lịch sử toàn bộ phím đã nhập ("bộ nhớ" của engine).
// inputMethod: kiểu gõ hiện tại (Telex, VNI...) đã parse từ file cấu hình.
// flags: cấu hình bật/tắt các tính năng (EfreeToneMarking, EstdToneStyle...).
// ---------------------------------------------------------------------------
type BambooEngine struct {
	composition []*Transformation
	inputMethod InputMethod
	flags       uint
}

// NewEngine khởi tạo một BambooEngine mới.
//
// Input: inputMethod — kiểu gõ đã parse; flag — cấu hình engine.
// Output: IEngine (thực chất là *BambooEngine) với composition rỗng.
func NewEngine(inputMethod InputMethod, flag uint) IEngine {
	engine := BambooEngine{
		inputMethod: inputMethod,
		flags:       flag,
	}
	return &engine
}

// GetInputMethod trả về kiểu gõ hiện tại của engine.
//
// Output: InputMethod đã truyền vào lúc NewEngine.
func (e *BambooEngine) GetInputMethod() InputMethod {
	return e.inputMethod
}

// SetFlag cập nhật cờ cấu hình engine.
//
// Input: flag — cờ mới (EfreeToneMarking | EstdToneStyle | ...).
func (e *BambooEngine) SetFlag(flag uint) {
	e.flags = flag
}

// GetFlag trả về cờ cấu hình hiện tại.
//
// Output: giá trị flags đang lưu trong engine.
func (e *BambooEngine) GetFlag(flag uint) uint {
	return e.flags
}

// IsValid kiểm tra xem âm tiết cuối cùng trong composition có hợp lệ không.
//
// Input: inputIsFullComplete — nếu true thì kiểm tra cả âm tiết hoàn chỉnh
//
//	(không chỉ tiềm năng).
//
// Output: true nếu âm tiết cuối là tiếng Việt hợp lệ.
// Logic: Gọi extractLastWord để tách âm tiết cuối, sau đó gọi isValid() kiểm tra.
// Ví dụ: "nguyeenx" → false; "nguyen" → true.
func (e *BambooEngine) IsValid(inputIsFullComplete bool) bool {
	var _, last = extractLastWord(e.composition, e.GetInputMethod().Keys)
	return isValid(last, inputIsFullComplete)
}

// GetProcessedString lấy chuỗi đã xử lý (tiếng Việt) từ composition.
//
// Input: mode — chế độ định dạng output (VietnameseMode, ToneLess, FullText...).
// Output: Chuỗi tiếng Việt đã Flatten.
// Logic:
//   - FullText: Flatten toàn bộ composition.
//   - PunctuationMode: tách âm tiết cuối kèm dấu câu rồi Flatten.
//   - Mặc định: tách âm tiết cuối (extractLastWord) rồi Flatten.
//
// Đây là hàm C++ gọi để lấy chuỗi hiển thị trên preedit.
func (e *BambooEngine) GetProcessedString(mode Mode) string {
	var tmp []*Transformation
	if mode&FullText != 0 {
		tmp = e.composition
	} else if mode&PunctuationMode != 0 {
		_, tmp = extractLastWordWithPunctuationMarks(e.composition, e.inputMethod.Keys)
		return Flatten(tmp, VietnameseMode)
	} else {
		_, tmp = extractLastWord(e.composition, e.inputMethod.Keys)
	}
	return Flatten(tmp, mode)
}

// getApplicableRules tìm tất cả Rule trong kiểu gõ có Key khớp với phím nhấn.
//
// Input: key — phím người dùng nhấn.
// Output: Danh sách Rule khả dụng (có thể rỗng).
// Logic: Duyệt e.inputMethod.Rules, so sánh Rule.Key == unicode.ToLower(key).
// Ví dụ: Kiểu Telex, nhấn 'a' → tìm các rule có Key == 'a'.
func (e *BambooEngine) getApplicableRules(key rune) []Rule {
	var applicableRules []Rule
	for _, inputRule := range e.inputMethod.Rules {
		if inputRule.Key == unicode.ToLower(key) {
			applicableRules = append(applicableRules, inputRule)
		}
	}
	return applicableRules
}

// findTargetByKey tìm transformation đích và rule sẽ áp dụng cho một phím.
//
// Input: composition hiện tại; key — phím mới.
// Output: (target, rule) — target là transformation bị tác động, rule là rule sẽ áp dụng.
// Logic: Gọi findTarget() (trong bamboo_utils.go) để tìm vị trí đặt dấu/thanh
//
//	hợp lý nhất trong âm tiết.
func (e *BambooEngine) findTargetByKey(composition []*Transformation, key rune) (*Transformation, Rule) {
	return findTarget(composition, e.getApplicableRules(key), e.flags)
}

// CanProcessKey kiểm tra xem phím có thuộc bảng gõ của kiểu gõ hiện tại không.
//
// Input: key — phím cần kiểm tra.
// Output: true nếu phím này nằm trong bảng phím của kiểu gõ.
// Logic: Gọi canProcessKey() (trong bamboo_utils.go).
// Ý nghĩa: C++ gọi trước để quyết định có đưa phím vào engine hay xử lý như phím thường.
func (e *BambooEngine) CanProcessKey(key rune) bool {
	return canProcessKey(key, e.inputMethod.Keys)
}

// generateTransformations tạo ra các transformation cho một phím mới.
//
// Input: composition hiện tại (thường là âm tiết cuối); lowerKey — phím thường hóa;
//
//	isUpperCase — phím gốc có phải chữ hoa không.
//
// Output: Danh sách transformation mới cho phím này.
// Logic:
//  1. Thử áp dụng các rule khả dụng qua generateTransformations() global.
//  2. Nếu không có rule nào áp dụng được → tạo fallback APPENDING transformation.
//  3. Kiểm tra shortcut uwo+ (ví dụ: uo + w → ươ). Nếu khớp, tạo virtual MARK_HORN.
//  4. Gọi refreshLastToneTarget() để điều chỉnh lại vị trí dấu thanh nếu cần.
//
// Đây là hàm cốt lõi nhất — quyết định mỗi phím tạo ra transformation gì.
func (e *BambooEngine) generateTransformations(composition []*Transformation, lowerKey rune, isUpperCase bool) []*Transformation {
	var transformations = generateTransformations(composition, e.getApplicableRules(lowerKey), e.flags, lowerKey, isUpperCase)
	if transformations == nil {
		// If none of the applicable_rules can actually be applied then this new
		// transformation fall-backs to an APPENDING one.
		transformations = generateFallbackTransformations(composition, e.getApplicableRules(lowerKey), lowerKey, isUpperCase)
		var newComposition = append(composition, transformations...)

		// Implement the uwo+ typing shortcut by creating a virtual
		// Mark.HORN rule that targets 'u' or 'o'.
		if virtualTrans := e.applyUowShortcut(newComposition); virtualTrans != nil {
			transformations = append(transformations, virtualTrans)
		}
	}
	/**
	* Sometimes, a tone's position in a previous state must be changed to fit the new state
	*
	* e.g.
	* prev state: chuyr -> chuỷ
	* this state: chuyrene -> chuyển
	**/
	transformations = append(transformations, e.refreshLastToneTarget(append(composition, transformations...))...)
	return transformations
}

// newComposition xây dựng composition mới sau khi xử lý một phím.
//
// Input: composition hiện tại; key; isUpperCase.
// Output: composition mới.
// Logic:
//  1. Tách composition thành previousTransformations (trước âm tiết cuối) và lastSyllable.
//  2. Gọi generateTransformations trên lastSyllable.
//  3. Nối lại: previousTransformations + lastSyllable mới.
//
// Ý nghĩa: Đảm bảo engine chỉ xử lý phím trên âm tiết cuối, không ảnh hưởng
//
//	các âm tiết trước đó.
func (e *BambooEngine) newComposition(composition []*Transformation, key rune, isUpperCase bool) []*Transformation {
	// Just process the key stroke on the last syllable
	var previousTransformations, lastSyllable = extractLastSyllable(composition)

	// Find all possible transformations this keypress can generate
	lastSyllable = append(lastSyllable, e.generateTransformations(lastSyllable, key, isUpperCase)...)

	// Put these transformations back to the composition
	return append(previousTransformations, lastSyllable...)
}

// applyUowShortcut tạo transformation ảo cho shortcut gõ nhanh uơ/ươ.
//
// Input: Một âm tiết (slice transformations).
// Output: Virtual transformation nếu khớp shortcut; nil nếu không.
// Logic:
//  1. Flatten âm tiết dạng ToneLess|LowerCase (ví dụ: "uo").
//  2. Kiểm tra regex regUOhTail (có kết thúc bằng uo/uơ...).
//  3. Nếu khớp, tìm target là chữ u hoặc o, tạo virtual MARK_HORN transformation.
//  4. Đặt Key = rune(0) để virtual rule không xuất hiện trong raw string.
//
// Ví dụ: gõ "uwo" → tự động thêm dấu móc vào "ươ".
func (e *BambooEngine) applyUowShortcut(syllable []*Transformation) *Transformation {
	str := Flatten(syllable, ToneLess|LowerCase)
	if len(e.inputMethod.SuperKeys) > 0 && regUOhTail.MatchString(str) {
		if target, missingRule := e.findTargetByKey(syllable, e.inputMethod.SuperKeys[0]); target != nil {
			missingRule.Key = rune(0) // virtual rule should not appear in the raw string
			virtualTrans := &Transformation{
				Rule:   missingRule,
				Target: target,
			}
			return virtualTrans
		}
	}
	return nil
}

// refreshLastToneTarget điều chỉnh lại vị trí dấu thanh nếu cần.
//
// Input: Một âm tiết.
// Output: Các transformation bổ sung để điều chỉnh vị trí dấu thanh.
// Logic: Nếu EfreeToneMarking bật và âm tiết hợp lệ, gọi refreshLastToneTarget()
//
//	global (trong bamboo_utils.go) để di chuyển dấu thanh sang vị trí đúng.
//
// Ví dụ: "chuyr" → "chuỷ" rồi "chuyrene" → "chuyển", dấu thanh chuyển từ y sang e.
func (e *BambooEngine) refreshLastToneTarget(syllable []*Transformation) []*Transformation {
	if e.flags&EfreeToneMarking != 0 && isValid(syllable, false) {
		return refreshLastToneTarget(syllable, e.flags&EstdToneStyle != 0)
	}
	return nil
}

/***** BEGIN SIDE-EFFECT METHODS ******/

// ProcessString xử lý một chuỗi ký tự bằng cách gọi ProcessKey cho từng ký tự.
//
// Input: str — chuỗi cần xử lý; mode — chế độ gõ.
// Logic: Lặp từng rune trong str, gọi e.ProcessKey.
// Ý nghĩa: Dùng khi paste chuỗi hoặc khôi phục trạng thái.
func (e *BambooEngine) ProcessString(str string, mode Mode) {
	for _, key := range str {
		e.ProcessKey(key, mode)
	}
}

// ProcessKey là entry point xử lý MỖI PHÍM người dùng nhấn.
//
// Input: key — phím vừa nhấn; mode — chế độ hiện tại.
// Logic:
//  1. Chuyển key về chữ thường (lowerKey), kiểm tra isUpperCase.
//  2. Nếu EnglishMode HOẶC !CanProcessKey(lowerKey):
//     - Phím không thuộc bảng gõ → thêm trực tiếp vào composition.
//     - Nếu InReverseOrder: thêm vào đầu; ngược lại: thêm vào cuối.
//  3. Ngược lại: gọi newComposition để xử lý tiếng Việt.
//
// Ý nghĩa: Đây là hàm C++ gọi qua CGO mỗi khi người dùng nhấn một phím.
func (e *BambooEngine) ProcessKey(key rune, mode Mode) {
	var lowerKey = unicode.ToLower(key)
	var isUpperCase = unicode.IsUpper(key)
	if mode&EnglishMode != 0 || !e.CanProcessKey(lowerKey) {
		if mode&InReverseOrder != 0 {
			e.composition = append([]*Transformation{newAppendingTrans(lowerKey, isUpperCase)}, e.composition...)
			return
		}
		e.composition = append(e.composition, newAppendingTrans(lowerKey, isUpperCase))
		return
	}
	e.composition = e.newComposition(e.composition, lowerKey, isUpperCase)
}

// RestoreLastWord khôi phục hoặc chuyển đổi âm tiết cuối giữa tiếng Việt và phím gốc.
//
// Input: toVietnamese — nếu true thì chuyển lại tiếng Việt; false thì tách rời phím gốc.
// Logic:
//   - Tách âm tiết cuối (extractLastWord).
//   - Nếu !toVietnamese: chuyển âm tiết cuối thành dạng "phím gốc" (breakComposition).
//   - Nếu toVietnamese: gõ lại âm tiết từ đầu bằng newComposition.
//
// Ý nghĩa: Dùng cho tính năng hoàn tác từ hoặc chuyển đổi tiếng Anh/Việt cho từ vừa gõ.
func (e *BambooEngine) RestoreLastWord(toVietnamese bool) {
	var previous, lastComb = extractLastWord(e.composition, e.GetInputMethod().Keys)
	if len(lastComb) == 0 {
		return
	}
	if !toVietnamese {
		e.composition = append(previous, breakComposition(lastComb)...)
	} else {
		var newComp []*Transformation
		for _, tnx := range lastComb {
			newComp = e.newComposition(newComp, tnx.Rule.Key, tnx.IsUpperCase)
		}
		e.composition = append(previous, newComp...)
	}
}

// Reset xóa toàn bộ trạng thái gõ.
//
// Logic: Gán e.composition = nil, xóa sạch lịch sử phím đã nhập.
// Ý nghĩa: Dùng khi commit preedit (nhấn Space/Enter) hoặc người dùng nhấn Escape.
func (e *BambooEngine) Reset() {
	e.composition = nil
}

// RemoveLastChar xóa ký tự cuối cùng (xử lý Backspace).
//
// Input: refreshLastToneTarget — có cần tính lại vị trí dấu thanh sau khi xóa không.
// Logic:
//  1. Tìm transformation APPENDING cuối cùng.
//  2. Nếu không tìm thấy → return.
//  3. Nếu phím đó không thuộc bảng gõ → chỉ cắt bớt 1 phần tử cuối.
//  4. Ngược lại: tách âm tiết cuối, lọc bỏ transformation cuối và các transformation
//     tác động lên nó.
//  5. Nếu refreshLastToneTarget thì tính lại dấu thanh.
//  6. Gán lại composition.
//
// Ý nghĩa: Không đơn giản là xóa 1 ký tự — phải xóa cả các dấu/mark gắn vào ký tự đó.
func (e *BambooEngine) RemoveLastChar(refreshLastToneTarget bool) {
	// Find the last APPENDING transformation and all
	// the transformations that add effects to it.
	var lastAppending = findLastAppendingTrans(e.composition)
	if lastAppending == nil {
		return
	}
	if !e.CanProcessKey(lastAppending.Rule.Key) {
		e.composition = e.composition[:len(e.composition)-1]
		return
	}
	var previous, lastComb = extractLastWord(e.composition, e.GetInputMethod().Keys)
	var newComb []*Transformation
	for _, t := range lastComb {
		if t.Target == lastAppending || t == lastAppending {
			continue
		}
		newComb = append(newComb, t)
	}
	if refreshLastToneTarget {
		newComb = append(newComb, e.refreshLastToneTarget(newComb)...)
	}
	e.composition = append(previous, newComb...)
}

/***** END SIDE-EFFECT METHODS ******/
