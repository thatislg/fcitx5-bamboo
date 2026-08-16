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
// File: input_method_def.go
// Mục đích: Định nghĩa các kiểu gõ tiếng Việt (Telex, Telex 2, Telex W).
//
// Kiến trúc:
//   - InputMethodDefinition: kiểu map[string]string, ánh xạ phím bấm -> hành động
//   - InputMethodDefinitions: biến toàn cục chứa tất cả các kiểu gõ
//   - GetInputMethodDefinitions(): trả về bản sao độc lập của định nghĩa
//
// Lưu ý:
//   - Đây là file DATA, không chứa logic xử lý phức tạp.
//   - Logic xử lý (parse rules, apply transformations) nằm trong rules_parser.go.
//   - Các hành động (XoaDauThanh, DauSac, A_Â...) sẽ được rules_parser.go
//     dịch thành các struct Transformation.
// =============================================================================

// InputMethodDefinition định nghĩa một kiểu gõ (ví dụ: Telex).
// Là map[string]string, trong đó:
//   - Key:   ký tự phím bấm (ví dụ: "s", "f", "w")
//   - Value: tên hành động cần thực hiện (ví dụ: "DauSac", "A_Â")
type InputMethodDefinition map[string]string

// InputMethodDefinitions chứa tất cả các kiểu gõ được hỗ trợ.
// Hiện tại còn 3 kiểu: Telex, Telex 2, Telex W.
//
// Mỗi kiểu gõ là một InputMethodDefinition, ánh xạ phím -> hành động.
// Ví dụ: "Telex" -> {"s": "DauSac", "f": "DauHuyen", ...}
//
// Các hành động có thể là:
//   - Thêm dấu thanh: DauSac, DauHuyen, DauHoi, DauNga, DauNang
//   - Xóa dấu thanh: XoaDauThanh
//   - Thêm dấu mũ/móc: A_Â, E_Ê, O_Ô
//   - Thêm dấu móc: UOA_ƯƠĂ (w -> ư/ơ/ă)
//   - Thêm dấu gạch ngang: D_Đ (dd -> đ)
var InputMethodDefinitions = map[string]InputMethodDefinition{
	"Telex": {
		"z": "XoaDauThanh", // Phím 'z': xóa dấu thanh (ví dụ: az -> a)
		"s": "DauSac",      // Phím 's': thêm dấu sắc (ví dụ: as -> á)
		"f": "DauHuyen",    // Phím 'f': thêm dấu huyền (ví dụ: af -> à)
		"r": "DauHoi",      // Phím 'r': thêm dấu hỏi (ví dụ: ar -> ả)
		"x": "DauNga",      // Phím 'x': thêm dấu ngã (ví dụ: ax -> ã)
		"j": "DauNang",     // Phím 'j': thêm dấu nặng (ví dụ: aj -> ạ)
		"a": "A_Â",         // Phím 'a': thêm dấu mũ (ví dụ: aa -> â)
		"e": "E_Ê",         // Phím 'e': thêm dấu mũ (ví dụ: ee -> ê)
		"o": "O_Ô",         // Phím 'o': thêm dấu mũ (ví dụ: oo -> ô)
		"w": "UOA_ƯƠĂ",     // Phím 'w': thêm dấu móc (ví dụ: uw -> ư)
		"d": "D_Đ",         // Phím 'd': thêm dấu gạch ngang (ví dụ: dd -> đ)
	},
	"Telex 2": {
		"z": "XoaDauThanh", // Phím 'z': xóa dấu thanh
		"s": "DauSac",      // Phím 's': thêm dấu sắc
		"f": "DauHuyen",    // Phím 'f': thêm dấu huyền
		"r": "DauHoi",      // Phím 'r': thêm dấu hỏi
		"x": "DauNga",      // Phím 'x': thêm dấu ngã
		"j": "DauNang",     // Phím 'j': thêm dấu nặng
		"a": "A_Â",         // Phím 'a': thêm dấu mũ
		"e": "E_Ê",         // Phím 'e': thêm dấu mũ
		"o": "O_Ô",         // Phím 'o': thêm dấu mũ
		"w": "UOA_ƯƠĂ__Ư",  // Phím 'w': thêm dấu móc (có thêm chữ hoa Ư)
		"d": "D_Đ",         // Phím 'd': thêm dấu gạch ngang
		"]": "__ư",         // Phím ']': thêm chữ thường ư (mở rộng)
		"[": "__ơ",         // Phím '[': thêm chữ thường ơ (mở rộng)
		"}": "_Ư",          // Phím '}': thêm chữ hoa Ư (mở rộng)
		"{": "_Ơ",          // Phím '{': thêm chữ hoa Ơ (mở rộng)
	},
	"Telex W": {
		"z": "XoaDauThanh", // Phím 'z': xóa dấu thanh
		"s": "DauSac",      // Phím 's': thêm dấu sắc
		"f": "DauHuyen",    // Phím 'f': thêm dấu huyền
		"r": "DauHoi",      // Phím 'r': thêm dấu hỏi
		"x": "DauNga",      // Phím 'x': thêm dấu ngã
		"j": "DauNang",     // Phím 'j': thêm dấu nặng
		"a": "A_Â",         // Phím 'a': thêm dấu mũ
		"e": "E_Ê",         // Phím 'e': thêm dấu mũ
		"o": "O_Ô",         // Phím 'o': thêm dấu mũ
		"w": "UOA_ƯƠĂ__Ư",  // Phím 'w': thêm dấu móc (có thêm chữ hoa Ư)
		"d": "D_Đ",         // Phím 'd': thêm dấu gạch ngang
	},
}

// GetInputMethodDefinitions trả về một bản sao (deep copy) của
// InputMethodDefinitions.
//
// Input:  không có
// Output: map[string]InputMethodDefinition — bản sao độc lập
//
// Mục đích: Bảo vệ biến toàn cục InputMethodDefinitions khỏi bị thay đổi
// bởi caller. rules_parser.go sẽ dùng hàm này để lấy dữ liệu.
//
// Lưu ý: Đây là shallow copy vì map trong Go là reference type,
// nhưng các giá trị string bên trong là immutable nên không có vấn đề.
func GetInputMethodDefinitions() map[string]InputMethodDefinition {
	var t = make(map[string]InputMethodDefinition)
	for k, v := range InputMethodDefinitions {
		t[k] = v
	}
	return t
}
