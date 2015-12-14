/*
Licensed to the Apache Software Foundation (ASF) under one
or more contributor license agreements.  See the NOTICE file
distributed with this work for additional information
regarding copyright ownership.  The ASF licenses this file
to you under the Apache License, Version 2.0 (the
"License"); you may not use this file except in compliance
with the License.  You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing,
software distributed under the License is distributed on an
"AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
KIND, either express or implied.  See the License for the
specific language governing permissions and limitations
under the License.
*/

package crypto

import (
	"crypto/ecdsa"
	"crypto/x509"
	"github.com/openblockchain/obc-peer/openchain/crypto/utils"
	obc "github.com/openblockchain/obc-peer/protos"
)

// Public Struct

type validatorImpl struct {
	peer *peerImpl

	isInitialized bool

	enrollCerts map[string]*x509.Certificate
}

func (validator *validatorImpl) GetName() string {
	return validator.peer.GetName()
}

// GetID returns this validator's identifier
func (validator *validatorImpl) GetID() []byte {
	return validator.peer.GetID()
}

// GetEnrollmentID returns this validator's enroolment id
func (validator *validatorImpl) GetEnrollmentID() string {
	return validator.peer.GetEnrollmentID()
}

// TransactionPreValidation verifies that the transaction is
// well formed with the respect to the security layer
// prescriptions (i.e. signature verification).
func (validator *validatorImpl) TransactionPreValidation(tx *obc.Transaction) (*obc.Transaction, error) {
	if !validator.isInitialized {
		return nil, utils.ErrNotInitialized
	}

	return validator.peer.TransactionPreValidation(tx)
}

// TransactionPreValidation verifies that the transaction is
// well formed with the respect to the security layer
// prescriptions (i.e. signature verification). If this is the case,
// the method prepares the transaction to be executed.
func (validator *validatorImpl) TransactionPreExecution(tx *obc.Transaction) (*obc.Transaction, error) {
	if !validator.isInitialized {
		return nil, utils.ErrNotInitialized
	}

	return tx, nil
}

// Sign signs msg with this validator's signing key and outputs
// the signature if no error occurred.
func (validator *validatorImpl) Sign(msg []byte) ([]byte, error) {
	return validator.signWithEnrollmentKey(msg)
}

// Verify checks that signature if a valid signature of message under vkID's verification key.
// If the verification succeeded, Verify returns nil meaning no error occurred.
// If vkID is nil, then the signature is verified against this validator's verification key.
func (validator *validatorImpl) Verify(vkID, signature, message []byte) error {
	cert, err := validator.getEnrollmentCert(vkID)
	if err != nil {
		validator.peer.node.log.Error("Failed getting enrollment cert for [%s]: %s", utils.EncodeBase64(vkID), err)
	}

	vk := cert.PublicKey.(*ecdsa.PublicKey)

	ok, err := validator.verify(vk, message, signature)
	if err != nil {
		validator.peer.node.log.Error("Failed verifying signature for [%s]: %s", utils.EncodeBase64(vkID), err)
	}

	if !ok {
		validator.peer.node.log.Error("Failed invalid signature for [%s]", utils.EncodeBase64(vkID))

		return utils.ErrInvalidSignature
	}

	return nil
}

// Private Methods

func (validator *validatorImpl) register(id string, pwd []byte, enrollID, enrollPWD string) error {
	if validator.isInitialized {
		log.Error("Registering [%s]...done! Initialization already performed", enrollID)

		return nil
	}

	// Register node
	peer := new(peerImpl)
	if err := peer.register("validator", id, pwd, enrollID, enrollPWD); err != nil {
		return err
	}
	validator.peer = peer
	validator.isInitialized = true
	return nil
}

func (validator *validatorImpl) init(name string, pwd []byte) error {
	if validator.isInitialized {
		validator.peer.node.log.Error("Already initializaed.")

		return nil
	}

	// Register node
	peer := new(peerImpl)
	if err := peer.init("validator", name, pwd); err != nil {
		return err
	}
	validator.peer = peer

	// Initialize keystore
	validator.peer.node.log.Info("Init keystore...")
	err := validator.initKeyStore()
	if err != nil {
		if err != utils.ErrKeyStoreAlreadyInitialized {
			validator.peer.node.log.Error("Keystore already initialized.")
		} else {
			validator.peer.node.log.Error("Failed initiliazing keystore %s", err)

			return err
		}
	}
	validator.peer.node.log.Info("Init keystore...done.")

	// Init crypto engine
	err = validator.initCryptoEngine()
	if err != nil {
		validator.peer.node.log.Error("Failed initiliazing crypto engine %s", err)
		return err
	}

	// initialized
	validator.isInitialized = true

	peer.node.log.Info("Initialization...done.")

	return nil
}

func (validator *validatorImpl) initCryptoEngine() error {
	validator.enrollCerts = make(map[string]*x509.Certificate)
	return nil
}

func (validator *validatorImpl) close() error {
	return validator.peer.close()
}
