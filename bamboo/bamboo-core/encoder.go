/*
 * Bamboo - A Vietnamese Input method editor
 * Copyright (C) Luong Thanh Lam <ltlam93@gmail.com>
 *
 * This software is licensed under the MIT license. For more information,
 * see <https://github.com/BambooEngine/bamboo-core/blob/master/LICENSE>.
 */

package bamboo

// =============================================================================
// CHÚ THÍCH THÊM (Tiếng Việt)
// =============================================================================
// File: encoder.go
// Mục đích: Mã hóa text tiếng Việt từ Unicode sang các charset khác (TCVN3, VNI, UTF-8...).
//
// Luồng xử lý:
//   Encode(charsetName, input)
//     ├── Nếu charsetName == "Unicode" → trả về input nguyên vẹn
//     ├── Nếu charsetName không tồn tại → trả về input nguyên vẹn (fallback)
//     └── Nếu tồn tại → lặp từng rune, tra bảng charsetDefinitions
//
// Ví dụ:
//   Encode("Unicode", "tiếng")         → "tiếng"
//   Encode("TCVN3 (ABC)", "tiếng")    → "tiÕng"
//   Encode("VNI Windows", "tiếng")    → "tieing"
// =============================================================================

// UNICODE là tên của charset mặc định (không cần chuyển đổi).
// Được dùng trong BambooMintConfig và GetCharsetNames().
const UNICODE = "Unicode"

// Encode chuyển đổi chuỗi Unicode sang charset đích.
//
// Input:
//   - charsetName: tên charset đích (ví dụ: "TCVN3 (ABC)", "VNI Windows", "Unicode")
//   - input: chuỗi Unicode đầu vào
//
// Output: chuỗi đã được mã hóa theo charset đích
//
// Logic:
//  1. Nếu charsetName == "Unicode" → trả về input nguyên vẹn
//  2. Tra bảng charsetDefinitions[charsetName]
//  3. Nếu không có → trả về input nguyên vẹn (fallback)
//  4. Lặp qua từng rune trong input:
//     - Nếu rune có trong bảng → output += giá trị trong charset
//     - Nếu không có → output += rune gốc (giữ nguyên ký tự không tiếng Việt)
//
// Ví dụ:
//
//	Encode("TCVN3 (ABC)", "tiếng") → "tiÕng"
//	Encode("KhôngTồnTại", "abc")    → "abc" (fallback)
func Encode(charsetName string, input string) string {
	if charsetName == UNICODE {
		return input
	}
	var output string
	if charset, found := charsetDefinitions[charsetName]; found {
		for _, chr := range input {
			if out, found := charset[chr]; found {
				output = output + out
			} else {
				output = output + string(chr)
			}
		}
	} else {
		output = input
	}
	return output
}

// GetCharsetNames trả về danh sách tất cả các charset được hỗ trợ.
//
// Output: []string — danh sách tên charset
//
// Logic:
//  1. Thêm "Unicode" vào đầu danh sách
//  2. Lặp qua tất cả keys trong charsetDefinitions
//  3. Trả về danh sách kết quả
//
// Ví dụ kết quả: ["Unicode", "TCVN3 (ABC)", "UTF-8", "VNI Windows", "VIQR", ...]
func GetCharsetNames() []string {
	var names []string
	names = append(names, UNICODE)
	for cs := range charsetDefinitions {
		names = append(names, cs)
	}
	return names
}
