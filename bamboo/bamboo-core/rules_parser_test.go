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
// File: rules_parser_test.go
// Mục đích: Test suite cho parser rule — kiểm tra việc parse rule từ chuỗi
//           macro định nghĩa kiểu gõ (Telex). Bao gồm parse dấu thanh, dấu phụ,
//           appending rule, và appended rules (macro đi kèm như "__ươ").
// =============================================================================

package bamboo

import (
	"testing"
)

// TestParseToneRules kiểm tra parse rule dấu thanh từ macro.
//
// Cases: z + "XoaDauThanh" → ToneNone; x + "DauNga" → ToneTilde.
func TestParseToneRules(t *testing.T) {
	rules := ParseRules('z', "XoaDauThanh")
	if len(rules) != 1 || rules[0].EffectType != ToneTransformation || Tone(rules[0].Effect) != ToneNone {
		t.Errorf("Test parse None Rule. Got %v, expected %v", rules[0], Rule{
			Key:        'z',
			EffectType: ToneTransformation,
			Effect:     0,
		})
	}
	rules = ParseRules('x', "DauNga")
	if len(rules) != 1 || rules[0].EffectType != ToneTransformation || rules[0].GetTone() != ToneTilde {
		t.Errorf("Test parse None Rule. Got %v, expected %v", rules[0], Rule{
			Key:        'x',
			EffectType: ToneTransformation,
			Effect:     uint8(ToneTilde),
		})
	}
}

// TestParseTonelessRules kiểm tra parse rule dấu phụ (mark) và appending từ macro.
//
// Cases: d + "D_Đ" → MarkDash + Appending Đ; { + "_Ư" → Appending Ư;
//
//	w + "UOA_ƯƠĂ" → 33 mark rules; w + "UOA_ƯƠĂ__Ư" → 34 rules (thêm appending ư).
func TestParseTonelessRules(t *testing.T) {
	rules := ParseTonelessRules('d', "D_Đ")
	idx := 0
	if len(rules) != 2 || rules[idx].EffectType != MarkTransformation || rules[idx].Effect != uint8(MarkDash) || rules[idx].EffectOn != 'd' {
		t.Errorf("Test parsing Mark Rule. Got %v, expected %v", rules[idx], Rule{
			Key:        'd',
			EffectType: MarkTransformation,
			Effect:     uint8(MarkDash),
			EffectOn:   'd',
		})
	}
	rules = ParseTonelessRules('{', "_Ư")
	if len(rules) != 1 || rules[0].EffectType != Appending || rules[0].EffectOn != 'Ư' {
		t.Errorf("Test parsing Append Rule. Got %v, expected %v", rules, Rule{
			Key:        '{',
			EffectType: Appending,
			EffectOn:   'Ư',
		})
	}
	rules = ParseTonelessRules('w', "UOA_ƯƠĂ")
	t.Log("RULES=", rules)
	if len(rules) != 33 {
		t.Errorf("Test the length of parsing mark rule. Got %d, expected %d", len(rules), 30)
	}
	if rules[0].EffectType != MarkTransformation || rules[0].GetMark() != MarkHorn || rules[0].EffectOn != 'u' {
		t.Errorf("Test parsing mark Rule. Got %v, expected %v", rules[0], Rule{
			Key:        'w',
			EffectType: MarkTransformation,
			Effect:     uint8(MarkHorn),
			EffectOn:   'u',
		})
	}
	idx = 7
	if rules[idx].EffectType != MarkTransformation || rules[idx].GetMark() != MarkHorn || rules[idx].EffectOn != 'o' {
		t.Errorf("Test parsing mark Rule. Got %v, expected %v", rules[idx], Rule{
			Key:        'w',
			EffectType: MarkTransformation,
			Effect:     uint8(MarkHorn),
			EffectOn:   'o',
		})
	}
	idx = 20
	if rules[idx].EffectType != MarkTransformation || rules[idx].GetMark() != MarkBreve || rules[idx].EffectOn != 'a' {
		t.Errorf("Test parsing mark Rule. Got %v, expected %v", rules[idx], Rule{
			Key:        'w',
			EffectType: MarkTransformation,
			Effect:     uint8(MarkBreve),
			EffectOn:   'a',
		})
	}
	rules = ParseTonelessRules('w', "UOA_ƯƠĂ__Ư")
	if len(rules) != 34 {
		t.Errorf("Test the length of parsing mark rule. Got %d, expected %d", len(rules), 31)
	} else {
		t.Log("RULES[UOA_ƯƠĂ__Ư]=", rules)
		idx = 20
		if rules[idx].EffectType != MarkTransformation || rules[idx].GetMark() != MarkBreve || rules[idx].EffectOn != 'a' {
			t.Errorf("Test parsing mark Rule. Got %v, expected %v", rules[idx], Rule{
				Key:        'w',
				EffectType: MarkTransformation,
				Effect:     uint8(MarkBreve),
				EffectOn:   'a',
			})
		}
		idx = 33
		if rules[idx].EffectType != Appending || rules[idx].EffectOn != 'ư' {
			t.Errorf("Test parsing mark Rule. Got %v, expected %v", rules[idx], Rule{
				Key:        'w',
				EffectType: Appending,
				EffectOn:   'ư',
			})
		}
	}

}

// TestAppendRule kiểm tra parse rule có AppendedRules (macro đi kèm như __ươ → append ơ).
//
// Cases: [ + "__ươ" → 1 rule với appended rule Appending ơ;
//
//	{ + "__ƯƠ" → 1 rule với appended rule Appending Ơ.
func TestAppendRule(t *testing.T) {
	rules := ParseTonelessRules('[', "__ươ")
	if len(rules) != 1 {
		t.Errorf("Test the length of parsing mark rule. Got %d, expected %d", len(rules), 1)
	} else {
		appendRules := rules[0].AppendedRules
		if len(appendRules) != 1 || appendRules[0].EffectType != Appending || appendRules[0].EffectOn != 'ơ' {
			t.Errorf("Test parsing append mark Rule. Got %v, expected %v", appendRules, Rule{
				Key:        '[',
				EffectType: Appending,
				EffectOn:   'ơ',
			})
		}
	}

	rules = ParseTonelessRules('{', "__ƯƠ")
	if len(rules) != 1 {
		t.Errorf("Test the length of parsing mark rule. Got %d, expected %d", len(rules), 1)
	} else {
		appendRules := rules[0].AppendedRules
		if len(appendRules) != 1 || appendRules[0].EffectType != Appending || appendRules[0].EffectOn != 'Ơ' {
			t.Errorf("Test parsing append mark Rule. Got %v, expected %v", appendRules, Rule{
				Key:        '{',
				EffectType: Appending,
				EffectOn:   'Ơ',
			})
		}
	}
}

// TestParseRulesWithIm (chưa implement).
func TestParseRulesWithIm(t *testing.T) {
}
