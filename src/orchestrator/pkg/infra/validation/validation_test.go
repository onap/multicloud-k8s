/*
 * Copyright 2018 Intel Corporation, Inc
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package validation

import (
	"bytes"
	"testing"
)

func TestIsTarGz(t *testing.T) {

	t.Run("Valid tar.gz", func(t *testing.T) {
		content := []byte{
			0x1f, 0x8b, 0x08, 0x08, 0xb0, 0x6b, 0xf4, 0x5b,
			0x00, 0x03, 0x74, 0x65, 0x73, 0x74, 0x2e, 0x74,
			0x61, 0x72, 0x00, 0xed, 0xce, 0x41, 0x0a, 0xc2,
			0x30, 0x10, 0x85, 0xe1, 0xac, 0x3d, 0x45, 0x4e,
			0x50, 0x12, 0xd2, 0xc4, 0xe3, 0x48, 0xa0, 0x01,
			0x4b, 0x52, 0x0b, 0xed, 0x88, 0x1e, 0xdf, 0x48,
			0x11, 0x5c, 0x08, 0xa5, 0x8b, 0x52, 0x84, 0xff,
			0xdb, 0xbc, 0x61, 0x66, 0x16, 0x4f, 0xd2, 0x2c,
			0x8d, 0x3c, 0x45, 0xed, 0xc8, 0x54, 0x21, 0xb4,
			0xef, 0xb4, 0x67, 0x6f, 0xbe, 0x73, 0x61, 0x9d,
			0xb2, 0xce, 0xd5, 0x55, 0xf0, 0xde, 0xd7, 0x3f,
			0xdb, 0xd6, 0x49, 0x69, 0xb3, 0x67, 0xa9, 0x8f,
			0xfb, 0x2c, 0x71, 0xd2, 0x5a, 0xc5, 0xee, 0x92,
			0x73, 0x8e, 0x43, 0x7f, 0x4b, 0x3f, 0xff, 0xd6,
			0xee, 0x7f, 0xea, 0x9a, 0x4a, 0x19, 0x1f, 0xe3,
			0x54, 0xba, 0xd3, 0xd1, 0x55, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x1b, 0xbc, 0x00, 0xb5, 0xe8,
			0x4a, 0xf9, 0x00, 0x28, 0x00, 0x00,
		}

		err := IsTarGz(bytes.NewBuffer(content))
		if err != nil {
			t.Errorf("Error reading valid tar.gz file %s", err.Error())
		}
	})

	t.Run("Invalid tar.gz", func(t *testing.T) {
		content := []byte{
			0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0xff, 0xf2, 0x48, 0xcd,
		}

		err := IsTarGz(bytes.NewBuffer(content))
		if err == nil {
			t.Errorf("Error should NOT be nil")
		}
	})

	t.Run("Empty tar.gz", func(t *testing.T) {
		content := []byte{}
		err := IsTarGz(bytes.NewBuffer(content))
		if err == nil {
			t.Errorf("Error should NOT be nil")
		}
	})
}

func TestIsValidName(t *testing.T) {
	t.Run("Valid Names", func(t *testing.T) {
		validnames := []string{
			"abc123",
			"1_abc123.ONE",
			"0abcABC_-.5",
			"123456789012345678901234567890123456789012345678901234567890123", // max of 63 characters
		}
		for _, name := range validnames {
			errs := IsValidName(name)
			if len(errs) > 0 {
				t.Errorf("Valid name failed to pass: %v", name)
			}
		}
	})

	t.Run("Invalid Names", func(t *testing.T) {
		invalidnames := []string{
			"",               // empty
			"_abc123",        // starts with non-alphanum
			"-abc123",        // starts with non-alphanum
			".abc123",        // starts with non-alphanum
			"abc123-",        // ends with non-alphanum
			"abc123_",        // ends with non-alphanum
			"abc123.",        // ends with non-alphanum
			"1_abc-123.O=NE", // contains not allowed character
			"1_a/bc-123.ONE", // contains not allowed character
			"1234567890123456789012345678901234567890123456789012345678901234", // longer than 63 characters
		}
		for _, name := range invalidnames {
			errs := IsValidName(name)
			if len(errs) == 0 {
				t.Errorf("Invalid name passed: %v", name)
			}
		}
	})
}

func TestIsIpv4Cidr(t *testing.T) {
	t.Run("Valid IPv4 Cidr", func(t *testing.T) {
		validipv4cidr := []string{
			"1.2.3.4/32",
			"10.11.12.0/24",
			"192.168.1.2/8",
			"255.0.0.0/16",
		}
		for _, ip := range validipv4cidr {
			err := IsIpv4Cidr(ip)
			if err != nil {
				t.Errorf("Valid IPv4 CIDR string failed to pass: %v", ip)
			}
		}
	})

	t.Run("Invalid IPv4 Cidr", func(t *testing.T) {
		invalidipv4cidr := []string{
			"",
			"1.2.3.4.5/32",
			"1.2.3.415/16",
			"1.2.3.4/33",
			"2.3.4/24",
			"1.2.3.4",
			"1.2.3.4/",
		}
		for _, ip := range invalidipv4cidr {
			err := IsIpv4Cidr(ip)
			if err == nil {
				t.Errorf("Invalid IPv4 Cidr passed: %v", ip)
			}
		}
	})
}

func TestIsIpv4(t *testing.T) {
	t.Run("Valid IPv4", func(t *testing.T) {
		validipv4 := []string{
			"1.2.3.42",
			"10.11.12.0",
			"192.168.1.2",
			"255.0.0.0",
			"255.255.255.255",
			"0.0.0.0",
		}
		for _, ip := range validipv4 {
			err := IsIpv4(ip)
			if err != nil {
				t.Errorf("Valid IPv4 string failed to pass: %v", ip)
			}
		}
	})

	t.Run("Invalid IPv4", func(t *testing.T) {
		invalidipv4 := []string{
			"",
			"1.2.3.4.5",
			"1.2.3.45/32",
			"1.2.3.4a",
			"2.3.4",
			"1.2.3.400",
			"256.255.255.255",
			"10,11,12,13",
			"1.2.3.4/",
		}
		for _, ip := range invalidipv4 {
			err := IsIpv4(ip)
			if err == nil {
				t.Errorf("Invalid IPv4 passed: %v", ip)
			}
		}
	})
}

func TestIsMac(t *testing.T) {
	t.Run("Valid MAC", func(t *testing.T) {
		validmacs := []string{
			"11:22:33:44:55:66",
			"ab-cd-ef-12-34-56",
			"AB-CD-EF-12-34-56",
		}
		for _, mac := range validmacs {
			err := IsMac(mac)
			if err != nil {
				t.Errorf("Valid MAC string failed to pass: %v", mac)
			}
		}
	})

	t.Run("Invalid MAC", func(t *testing.T) {
		invalidmacs := []string{
			"",
			"1.2.3.4.5",
			"1.2.3.45/32",
			"ab:cd:ef:gh:12:34",
			"11:22-33-44:55:66",
			"11,22,33,44,55,66",
			"11|22|33|44|55|66",
			"11:22:33:44:55:66:77",
			"11-22-33-44-55",
			"11-22-33-44-55-66-77",
		}
		for _, mac := range invalidmacs {
			err := IsMac(mac)
			if err == nil {
				t.Errorf("Invalid MAC passed: %v", mac)
			}
		}
	})
}

func TestIsValidString(t *testing.T) {
	t.Run("Valid Strings", func(t *testing.T) {
		validStrings := []struct {
			str    string
			min    int
			max    int
			format string
		}{
			{
				str:    "abc123",
				min:    0,
				max:    16,
				format: VALID_NAME_STR,
			},
			{
				str:    "ab-c1_2.3",
				min:    0,
				max:    16,
				format: VALID_NAME_STR,
			},
			{
				str:    "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ",
				min:    0,
				max:    62,
				format: VALID_ALPHANUM_STR,
			},
			{
				str:    "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ",
				min:    0,
				max:    52,
				format: VALID_ALPHA_STR,
			},
			{
				str:    "",
				min:    0,
				max:    52,
				format: VALID_ALPHA_STR,
			},
			{
				str:    "",
				min:    0,
				max:    52,
				format: VALID_ALPHANUM_STR,
			},
			{
				str:    "dGhpcyBpcyBhCnRlc3Qgc3RyaW5nCg==",
				min:    0,
				max:    52,
				format: VALID_BASE64_STR,
			},
			{
				str:    "\t\n \n0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ!@#$%^&*()_+-=,.<>/?'\"\\[]{}\n",
				min:    0,
				max:    256,
				format: VALID_ANY_STR,
			},
		}
		for _, test := range validStrings {
			errs := IsValidString(test.str, test.min, test.max, test.format)
			if len(errs) > 0 {
				t.Errorf("Valid string failed to pass: str:%v, min:%v, max:%v, format:%v", test.str, test.min, test.max, test.format)
			}
		}
	})

	t.Run("Invalid Strings", func(t *testing.T) {
		inValidStrings := []struct {
			str    string
			min    int
			max    int
			format string
		}{
			{
				str:    "abc123",
				min:    0,
				max:    5,
				format: VALID_NAME_STR,
			},
			{
				str:    "",
				min:    0,
				max:    5,
				format: VALID_NAME_STR,
			},
			{
				str:    "-ab-c1_2.3",
				min:    0,
				max:    16,
				format: VALID_NAME_STR,
			},
			{
				str:    "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ=",
				min:    0,
				max:    100,
				format: VALID_ALPHANUM_STR,
			},
			{
				str:    "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ123456",
				min:    0,
				max:    62,
				format: VALID_ALPHA_STR,
			},
			{
				str:    "",
				min:    1,
				max:    52,
				format: VALID_ALPHA_STR,
			},
			{
				str:    "abc123",
				min:    1,
				max:    3,
				format: VALID_ALPHA_STR,
			},
			{
				str:    "",
				min:    1,
				max:    52,
				format: VALID_ALPHANUM_STR,
			},
			{
				str:    "dGhpcyBpcyBhCnRlc3Qgc3RyaW5nCg===",
				min:    0,
				max:    52,
				format: VALID_BASE64_STR,
			},
			{
				str:    "dGhpcyBpcyBhCnRlc3=Qgc3RyaW5nCg==",
				min:    0,
				max:    52,
				format: VALID_BASE64_STR,
			},
			{
				str:    "dGhpcyBpcyBhCnRlc3#Qgc3RyaW5nCg==",
				min:    0,
				max:    52,
				format: VALID_BASE64_STR,
			},
			{
				str:    "\t\n \n0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ!@#$%^&*()_+-=,.<>/?'\"\\[]{}\n",
				min:    0,
				max:    10,
				format: VALID_ANY_STR,
			},
			{
				str:    "abc123",
				min:    0,
				max:    10,
				format: "unknownformat",
			},
		}
		for _, test := range inValidStrings {
			errs := IsValidString(test.str, test.min, test.max, test.format)
			if len(errs) == 0 {
				t.Errorf("Invalid string passed: str:%v, min:%v, max:%v, format:%v", test.str, test.min, test.max, test.format)
			}
		}
	})
}

func TestIsValidLabel(t *testing.T) {
	t.Run("Valid Labels", func(t *testing.T) {
		validlabels := []string{
			"kubernetes.io/hostname=localhost",
			"hostname=localhost",
			"localhost",
		}
		for _, label := range validlabels {
			errs := IsValidLabel(label)
			if len(errs) > 0 {
				t.Errorf("Valid label failed to pass: %v %v", label, errs)
			}
		}
	})

	t.Run("Invalid Labels", func(t *testing.T) {
		invalidlabels := []string{
			"",
			"kubernetes$.io/hostname=localhost",
			"hostname==localhost",
			"=localhost",
			"/hostname=localhost",
			".a.b/hostname=localhost",
			"kubernetes.io/hostname",
			"kubernetes.io/hostname=",
			"kubernetes.io/1234567890123456789012345678901234567890123456789012345678901234=localhost",         // too long name
			"kubernetes.io/hostname=localhost1234567890123456789012345678901234567890123456789012345678901234", // too long value
			"12345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234/hostname=localhost", // too long prefix
		}
		for _, label := range invalidlabels {
			errs := IsValidLabel(label)
			if len(errs) == 0 {
				t.Errorf("Invalid label passed: %v", label)
			}
		}
	})
}

func TestIsValidNumber(t *testing.T) {
	t.Run("Valid Number", func(t *testing.T) {
		validNumbers := []struct {
			value int
			min   int
			max   int
		}{
			{
				value: 0,
				min:   0,
				max:   5,
			},
			{
				value: 1000,
				min:   0,
				max:   4095,
			},
			{
				value: 0,
				min:   0,
				max:   0,
			},
			{
				value: -100,
				min:   -200,
				max:   -99,
			},
			{
				value: 123,
				min:   123,
				max:   123,
			},
		}
		for _, test := range validNumbers {
			err := IsValidNumber(test.value, test.min, test.max)
			if len(err) > 0 {
				t.Errorf("Valid number failed to pass - value:%v, min:%v, max:%v", test.value, test.min, test.max)
			}
		}
	})

	t.Run("Invalid Number", func(t *testing.T) {
		inValidNumbers := []struct {
			value int
			min   int
			max   int
		}{
			{
				value: 6,
				min:   0,
				max:   5,
			},
			{
				value: 4096,
				min:   0,
				max:   4095,
			},
			{
				value: 11,
				min:   10,
				max:   10,
			},
			{
				value: -100,
				min:   -99,
				max:   -200,
			},
			{
				value: 123,
				min:   223,
				max:   123,
			},
		}
		for _, test := range inValidNumbers {
			err := IsValidNumber(test.value, test.min, test.max)
			if len(err) == 0 {
				t.Errorf("Invalid number passed - value:%v, min:%v, max:%v", test.value, test.min, test.max)
			}
		}
	})
}
