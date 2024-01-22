/*
 * Copyright contributors to the Galasa project
 *
 * SPDX-License-Identifier: EPL-2.0
 */
package images

type TerminalField struct {
    Row 				int             `json:"row"`
    Column 				int             `json:"column"`
    Unformatted 		bool            `json:"unformatted"`
    FieldProtected 		bool            `json:"fieldProtected"`
    FieldNumeric 		bool            `json:"fieldNumeric"`
    FieldDisplay 		bool            `json:"fieldDisplay"`
    FieldIntenseDisplay bool            `json:"fieldIntenseDisplay"`
    FieldSelectorPen    bool            `json:"fieldSelectorPen"`
    FieldModified 		bool            `json:"fieldModified"`
    ForegroundColor 	string          `json:"foregroundColour"`
    BackgroundColor 	string          `json:"backgroundColour"`
    Highlight 			string          `json:"highlight"`
    Contents 			[]FieldContents `json:"contents"`
}