/*
 * Copyright (C) 2018 The ontology Authors
 * This file is part of The ontology library.
 *
 * The ontology is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The ontology is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The ontology.  If not, see <http://www.gnu.org/licenses/>.
 */

package password

import (
	"bytes"
	"fmt"
	"os"

	"github.com/howeyc/gopass"
)

// GetPassword gets password from user input
func GetPassword() ([]byte, error) {
	fmt.Printf("Password:")
	passwd, err := gopass.GetPasswd()
	if err != nil {
		return nil, err
	}
	return passwd, nil
}

// GetConfirmedPassword gets double confirmed password from user input
func GetConfirmedPassword() ([]byte, error) {
	fmt.Printf("Password:")
	first, err := gopass.GetPasswd()
	if err != nil {
		return nil, err
	}
	if 0 == len(first) {
		fmt.Println("User have to input password.")
		os.Exit(1)
	}

	fmt.Printf("Re-enter Password:")
	second, err := gopass.GetPasswd()
	if err != nil {
		return nil, err
	}
	if 0 == len(second) {
		fmt.Println("User have to input password.")
		os.Exit(1)
	}

	if !bytes.Equal(first, second) {
		fmt.Println("Unmatched Password")
		os.Exit(1)
	}
	return first, nil
}

// GetPassword gets node's wallet password from command line or user input
func GetAccountPassword() ([]byte, error) {
	var passwd []byte
	var err error
	passwd, err = GetPassword()
	if err != nil {
		return nil, err
	}

	return passwd, nil
}
