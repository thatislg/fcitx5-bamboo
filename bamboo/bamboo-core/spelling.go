/*
 * Bamboo - A Vietnamese Input method editor
 * Copyright (C) Luong Thanh Lam <ltlam93@gmail.com>
 *
 * This software is licensed under the MIT license. For more information,
 * see <https://github.com/BambooEngine/bamboo-core/blob/master/LICENCE>.
 */

package bamboo

// =============================================================================
// CHÚ THÍCH THÊM (Tiếng Việt)
// =============================================================================
// File: spelling.go
// Mục đích: Kiểm tra chính tả (spell check) cho tiếng Việt.
//
// Kiến trúc:
//   - firstConsonantSeqs: các nhóm phụ âm đầu (C)
//   - vowelSeqs: các nhóm nguyên âm (V)
//   - lastConsonantSeqs: các nhóm phụ âm cuối (C)
//   - cvMatrix: ma trận cho phép C + V
//   - vcMatrix: ma trận cho phép V + C
//
// Luồng kiểm tra:
//   isValidCVC(fc, vo, lc)
//     ├── lookup(fc) → nhóm phụ âm đầu
//     ├── lookup(vo) → nhóm nguyên âm
//     ├── lookup(lc) → nhóm phụ âm cuối
//     ├── isValidCV(fcGroups, voGroups) → kiểm tra cvMatrix
//     └── isValidVC(voGroups, lcGroups) → kiểm tra vcMatrix
//
// Ví dụ:
//   isValidCVC("tr", "ươ", "ng", true) → true  // "trường" hợp lệ
//   isValidCVC("k", "ỳ", "", true)     → false // "kỳ" không hợp lệ
// =============================================================================

// firstConsonantSeqs chứa các nhóm phụ âm đầu được phép đứng ở đầu âm tiết.
// Các phụ âm được nhóm theo tính chất âm vị học.
// Ví dụ: nhóm 0 gồm các phụ âm đơn giản (b, d, đ, g...).
var firstConsonantSeqs = []string{
	"b d đ g gh m n nh p ph r s t tr v z",
	"c h k kh qu th",
	"ch gi l ng ngh x",
	"đ l",
	"h",
}

// vowelSeqs chứa các nhóm nguyên âm.
// Nhóm 5 là các diphthong/triphthong (ai, ao, iêu...) không có phụ âm cuối.
// Nhóm 6 ("ă") và nhóm 7 ("i") chỉ đi với một số phụ âm đầu đặc biệt.
var vowelSeqs = []string{
	"ê i ua uê uy y",
	"a iê oa uyê yê",
	"â ă e o oo ô ơ oe u ư uâ uô ươ",
	"oă",
	"uơ",
	"ai ao au âu ay ây eo êu ia iêu iu oai oao oay oeo oi ôi ơi ưa uây ui ưi uôi ươi ươu ưu uya uyu yêu",
	"ă",
	"i",
}

// lastConsonantSeqs chứa các nhóm phụ âm cuối.
// Nhóm 0 (ch, nh) và nhóm 1 (c, ng) là phổ biến nhất.
var lastConsonantSeqs = []string{
	"ch nh",
	"c ng",
	"m n p t",
	"k",
	"c",
}

// cvMatrix định nghĩa các nhóm nguyên âm được phép đi sau mỗi nhóm phụ âm đầu.
//
// Ví dụ: cvMatrix[0] = {0, 1, 2, 5}
//
//	→ Phụ âm nhóm 0 (b, d, đ...) có thể đi với nguyên âm nhóm 0, 1, 2, 5.
//	→ "ba" (a ∈ nhóm 1) ✓, "bê" (ê ∈ nhóm 0) ✓, "bă" (ă ∈ nhóm 6) ✗
//
// Ví dụ: cvMatrix[3] = {6}
//
//	→ Phụ âm nhóm 3 (đ, l) chỉ đi với nguyên âm nhóm 6 (ă).
//	→ "đăng" ✓, "đi" (i ∈ nhóm 7) ✗
var cvMatrix = [][]int{
	{0, 1, 2, 5},
	{0, 1, 2, 3, 4, 5},
	{0, 1, 2, 3, 5},
	{6},
	{7},
}

// vcMatrix định nghĩa các nhóm phụ âm cuối được phép đi sau mỗi nhóm nguyên âm.
//
// Ví dụ: vcMatrix[5] = {}
//
//	→ Nguyên âm nhóm 5 (diphthong: ai, ao, iêu...) không có phụ âm cuối.
//	→ "mai" ✓, "may" ✓, "mang" ✗ (sai vì "ai" đã là nguyên âm đôi)
//
// Ví dụ: vcMatrix[0] = {0, 2}
//
//	→ Nguyên âm nhóm 0 (ê, i...) có thể đi với phụ âm cuối nhóm 0 (ch, nh) hoặc nhóm 2 (m, n, p, t).
//	→ "êch" ✓, "in" ✓
var vcMatrix = [][]int{
	{0, 2},
	{0, 1, 2},
	{1, 2},
	{1, 2},
	{},
	{},
	{3},
	{4},
}

// lookup tìm nhóm của một chuỗi (phụ âm, nguyên âm, hoặc phụ âm cuối)
// trong danh sách các nhóm.
//
// Input:
//   - seq: danh sách các nhóm (firstConsonantSeqs, vowelSeqs, hoặc lastConsonantSeqs)
//   - input: chuỗi cần tìm (ví dụ: "tr", "ươ", "ng")
//   - inputIsFull: có phải chuỗi đầy đủ không
//   - inputIsComplete: có phải đã nhập xong không
//
// Output: []int — các chỉ số nhóm mà input thuộc về
//
// Ví dụ:
//
//	lookup(firstConsonantSeqs, "tr", true, true) → [0]   // "tr" trong nhóm 0
//	lookup(vowelSeqs, "ươ", true, true)          → [2]   // "ươ" trong nhóm 2
func lookup(seq []string, input string, inputIsFull, inputIsComplete bool) []int {
	var ret []int
	var inputLen = len([]rune(input))
	for index, row := range seq {
		var i = 0
		var rows = append([]rune(row), ' ')
		for j, char := range rows {
			if char != ' ' {
				continue
			}
			var canvas = rows[i:j]
			i = j + 1
			if len(canvas) < inputLen || (inputIsFull && len(canvas) > inputLen) {
				continue
			}
			var isMatch = true
			for k, ic := range []rune(input) {
				if ic != canvas[k] && !(!inputIsComplete && AddMarkToTonelessChar(canvas[k], 0) == ic) {
					isMatch = false
					break
				}
			}
			if isMatch {
				ret = append(ret, index)
				break
			}
		}
	}
	return ret
}

// isValidCVC kiểm tra xem một âm tiết tiếng Việt có hợp lệ không.
//
// Input:
//   - fc: phụ âm đầu (ví dụ: "tr")
//   - vo: nguyên âm (ví dụ: "ươ")
//   - lc: phụ âm cuối (ví dụ: "ng")
//   - inputIsFullComplete: người dùng đã nhập xong chưa
//
// Output: bool — âm tiết có hợp lệ không
//
// Luồng kiểm tra:
//  1. Tìm nhóm của fc trong firstConsonantSeqs
//  2. Tìm nhóm của vo trong vowelSeqs
//  3. Tìm nhóm của lc trong lastConsonantSeqs
//  4. Kiểm tra cvMatrix: fc group có đi với vo group không
//  5. Kiểm tra vcMatrix: vo group có đi với lc group không
//
// Ví dụ:
//
//	isValidCVC("tr", "ươ", "ng", true) → true   // "trường" hợp lệ
//	isValidCVC("k", "ỳ", "", true)     → false  // "kỳ" không hợp lệ
func isValidCVC(fc, vo, lc string, inputIsFullComplete bool) bool {
	var ret bool
	var fcIndexes, voIndexes, lcIndexes []int
	// log.Printf("fc=%s vo=%s lc=%s ret=%v", fc, vo, lc, ret)
	if fc != "" {
		if fcIndexes = lookup(firstConsonantSeqs, fc, inputIsFullComplete || vo != "", true); fcIndexes == nil {
			return false
		}
	}
	if vo != "" {
		if voIndexes = lookup(vowelSeqs, vo, inputIsFullComplete || lc != "", inputIsFullComplete); voIndexes == nil {
			return false
		}
	}
	if lc != "" {
		if lcIndexes = lookup(lastConsonantSeqs, lc, inputIsFullComplete, true); lcIndexes == nil {
			return false
		}
	}
	if voIndexes == nil {
		// first consonant only
		return fcIndexes != nil
	}
	if fcIndexes != nil {
		// first consonant + vowel
		if ret = isValidCV(fcIndexes, voIndexes); !ret || lcIndexes == nil {
			return ret
		}
	}
	if lcIndexes != nil {
		// vowel + last consonant
		ret = isValidVC(voIndexes, lcIndexes)
	} else {
		// vowel only
		ret = true
	}
	return ret
}

func isValidCV(fcIndexes, voIndexes []int) bool {
	for _, fc := range fcIndexes {
		for _, c := range cvMatrix[fc] {
			for _, vo := range voIndexes {
				if c == vo {
					return true
				}
			}
		}
	}
	return false
}

func isValidVC(voIndexes, lcIndexes []int) bool {
	for _, vo := range voIndexes {
		for _, c := range vcMatrix[vo] {
			for _, lc := range lcIndexes {
				if c == lc {
					return true
				}
			}
		}
	}
	return false
}
