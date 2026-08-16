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
// File: flattener.go
// Mục đích: Chuyển đổi danh sách các Transformation (biến đổi từ mỗi phím bấm)
//            thành chuỗi ký tự hiển thị (preedit text).
//
// Luồng xử lý:
//   composition []*Transformation → Flatten(mode) → string (preedit)
//
// Ví dụ:
//   Phím: a, w, s
//   composition: [{a}, {w→ư}, {s→sắc}]
//   appendingList: [a, w]           // Các ký tự gốc (Appending)
//   appendingMap: {a: [w, s]}       // w và s tác động lên a
//   Kết quả: "ứ"
//
// Các Mode ảnh hưởng hiển thị:
//   - EnglishMode: Hiển thị phím gốc (aws → "aws")
//   - ToneLess:    Bỏ dấu thanh (ứ → "ư")
//   - MarkLess:    Bỏ dấu mũ/móc (ứ → "ú")
//   - LowerCase:   Chuyển thường (Ứ → "ứ")
// =============================================================================

// Flatten chuyển đổi composition thành chuỗi hiển thị.
//
// Input:
//   - composition: danh sách các Transformation từ mỗi phím bấm
//   - mode: chế độ hiển thị (EnglishMode, ToneLess, MarkLess...)
//
// Output: string — chuỗi preedit hiển thị cho người dùng
//
// Ví dụ:
//
//	Flatten([{a}, {w→ư}, {s→sắc}], VietnameseMode) → "ứ"
//	Flatten([{a}, {w→ư}, {s→sắc}], EnglishMode)    → "aws"
func Flatten(composition []*Transformation, mode Mode) string {
	return string(getCanvas(composition, mode))
}

// getCanvas xây dựng []rune từ composition theo mode.
//
// Luồng xử lý:
//
//  1. Phân loại Transformation:
//     - appendingList: các ký tự gốc (EffectType = Appending)
//     - appendingMap: các biến đổi (Mark/Tone) áp lên từng ký tự
//
//  2. Với mỗi ký tự trong appendingList:
//     - Áp dụng MarkTransformation (nếu có)
//     - Áp dụng ToneTransformation (nếu có)
//     - Xử lý ToneLess, MarkLess, LowerCase/UpperCase
//     - Thêm vào canvas
func getCanvas(composition []*Transformation, mode Mode) []rune {
	var canvas []rune
	var appendingMap = map[*Transformation][]*Transformation{}
	var appendingList []*Transformation
	for _, trans := range composition {
		if mode&EnglishMode != 0 {
			if trans.Rule.Key == 0 {
				// ignore virtual key
				continue
			}
			appendingList = append(appendingList, trans)
		} else if trans.Rule.EffectType == Appending {
			if trans.Rule.Key == 0 {
				// ignore virtual key
				continue
			}
			appendingList = append(appendingList, trans)
		} else if trans.Target != nil {
			appendingMap[trans.Target] = append(appendingMap[trans.Target], trans)
		}
	}
	for _, appendingTrans := range appendingList {
		var chr rune
		var transList = appendingMap[appendingTrans]
		if mode&EnglishMode != 0 {
			chr = appendingTrans.Rule.Key
		} else {
			chr = appendingTrans.Rule.EffectOn
			for _, trans := range transList {
				switch trans.Rule.EffectType {
				case MarkTransformation:
					if trans.Rule.Effect == uint8(MarkRaw) {
						chr = appendingTrans.Rule.Key
					} else {
						chr = AddMarkToChar(chr, trans.Rule.Effect)
					}
				case ToneTransformation:
					chr = AddToneToChar(chr, trans.Rule.Effect)
				}
			}
		}
		if mode&ToneLess != 0 {
			chr = AddToneToChar(chr, 0)
		}
		if mode&MarkLess != 0 {
			chr = AddMarkToChar(chr, 0)
		}
		if mode&LowerCase != 0 {
			chr = unicode.ToLower(chr)
		} else if appendingTrans.IsUpperCase {
			chr = unicode.ToUpper(chr)
		}
		canvas = append(canvas, chr)
	}
	return canvas
}
