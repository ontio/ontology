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

package validation

import (
	"errors"
	"fmt"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/constants"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/core/payload"
	"github.com/ontio/ontology/core/signature"
	"github.com/ontio/ontology/core/types"
	ontErrors "github.com/ontio/ontology/errors"
	"github.com/ontio/ontology/smartcontract/service/wasmvm"
)

func EnsureTransactionVerified(tx *types.Transaction, extraSignerListEnabled bool) error {
	if err := checkExtraSignersEnabled(tx, extraSignerListEnabled); err != nil {
		return fmt.Errorf("verify transaction failed with %s", err.Error())
	}

	verified := tx.GetVerifiedSignerAddresses()
	if len(verified) == 0 {
		e := VerifyTransaction(tx, extraSignerListEnabled)
		if e != ontErrors.ErrNoError {
			return fmt.Errorf("verify transaction failed with %s", e.Error())
		}
		return nil
	}

	return nil
}

// VerifyTransaction verifys received single transaction
func VerifyTransaction(tx *types.Transaction, extraSignerListEnabled bool) ontErrors.ErrCode {
	if err := checkTransactionSignatures(tx, extraSignerListEnabled); err != nil {
		log.Info("transaction verify error:", err)
		return ontErrors.ErrVerifySignature
	}

	if err := checkTransactionPayload(tx); err != nil {
		log.Warn("[VerifyTransaction],", err)
		return ontErrors.ErrTransactionPayload
	}

	return ontErrors.ErrNoError
}

func checkTransactionSignatures(tx *types.Transaction, extraSignerListEnabled bool) error {
	if tx.IsEipTx() {
		//the signature already checked on decode tx
		return nil
	}
	if err := checkExtraSignersEnabled(tx, extraSignerListEnabled); err != nil {
		return err
	}

	hash := tx.Hash()

	lensig := len(tx.Sigs)
	if lensig > constants.TX_MAX_SIG_SIZE {
		return fmt.Errorf("transaction signature number %d execced %d", lensig, constants.TX_MAX_SIG_SIZE)
	}

	address := make(map[common.Address]bool, len(tx.Sigs))
	for _, sig := range tx.Sigs {
		m := int(sig.M)
		kn := len(sig.PubKeys)
		sn := len(sig.SigData)

		if kn > constants.MULTI_SIG_MAX_PUBKEY_SIZE || sn < m || m > kn || m <= 0 {
			return errors.New("wrong tx sig param length")
		}

		if kn == 1 {
			err := signature.Verify(sig.PubKeys[0], hash[:], sig.SigData[0])
			if err != nil {
				return errors.New("signature verification failed")
			}
		} else {
			if err := signature.VerifyMultiSignature(hash[:], sig.PubKeys, m, sig.SigData); err != nil {
				return err
			}
		}
		address[sig.SignerAddress()] = true
	}

	if err := checkSignerSet(tx, address, extraSignerListEnabled); err != nil {
		return err
	}

	addrList := make([]common.Address, 0, len(address))
	for addr := range address {
		addrList = append(addrList, addr)
	}
	tx.SetSignedAddresses(addrList)

	return nil
}

// checkExtraSignersEnabled rejects extra signers before the feature is
// enabled. After activation, checkSignerSet validates the list against the
// verified signer set.
func checkExtraSignersEnabled(tx *types.Transaction, extraSignerListEnabled bool) error {
	if !extraSignerListEnabled && len(tx.ExtraSigners) > 0 {
		return errors.New("transaction extra signers are not enabled")
	}
	return nil
}

// checkSignerSet verifies that Payer is in the verified signer set and, when
// the extra signer rule is active, that ExtraSigners is exactly the verified
// set minus Payer. It performs no cryptographic work.
func checkSignerSet(tx *types.Transaction, address map[common.Address]bool, extraSignersEnabled bool) error {
	if !address[tx.Payer] {
		return errors.New("signature missing for payer: " + tx.Payer.ToBase58())
	}
	if !extraSignersEnabled {
		return nil
	}
	if len(address) != len(tx.ExtraSigners)+1 {
		return errors.New("wrong number of signers in the transaction")
	}
	for _, addr := range tx.ExtraSigners {
		if !address[addr] {
			return errors.New("signature missing for address: " + addr.ToBase58())
		}
	}
	return nil
}

func checkTransactionPayload(tx *types.Transaction) error {
	switch pld := tx.Payload.(type) {
	case *payload.DeployCode:
		deploy := tx.Payload.(*payload.DeployCode)
		if deploy.VmType() == payload.WASMVM_TYPE {
			_, err := wasmvm.ReadWasmModule(deploy.GetRawCode(), config.DefConfig.Common.WasmVerifyMethod)
			if err != nil {
				return err
			}
		}
		return nil
	case *payload.InvokeCode:
		return nil
	case *payload.EIP155Code:
		return nil
	default:
		return errors.New(fmt.Sprint("[txValidator], unimplemented transaction payload type.", pld))
	}
}
