package e2e

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/ory/dockertest/v3/docker"

	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

const (
	flagFrom            = "from"
	flagHome            = "home"
	flagFees            = "fees"
	flagGas             = "gas"
	flagOutput          = "output"
	flagChainID         = "chain-id"
	flagSpendLimit      = "spend-limit"
	flagGasAdjustment   = "gas-adjustment"
	flagFeeGranter      = "fee-granter"
	flagBroadcastMode   = "broadcast-mode"
	flagKeyringBackend  = "keyring-backend"
	flagAllowedMessages = "allowed-messages"
)

type flagOption func(map[string]interface{})

// withKeyValue add a new flag to command

func withKeyValue(key string, value interface{}) flagOption {
	return func(o map[string]interface{}) {
		o[key] = value
	}
}

func applyOptions(chainID string, options []flagOption) map[string]interface{} {
	opts := map[string]interface{}{
		flagKeyringBackend: "test",
		flagOutput:         "json",
		flagGas:            "auto",
		flagFrom:           "alice",
		flagBroadcastMode:  "sync",
		flagGasAdjustment:  "1.5",
		flagChainID:        chainID,
		flagHome:           onomyHomePath,
		flagFees:           standardFees.String(),
	}
	for _, apply := range options {
		apply(opts)
	}
	return opts
}

func (s *IntegrationTestSuite) execEncode(
	c *chain,
	txPath string,
	opt ...flagOption,
) string {
	opts := applyOptions(c.id, opt)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	s.T().Logf("%s - Executing onomyd encoding with %v", c.id, txPath)
	onomyCommand := []string{
		onomydBinary,
		txCommand,
		"encode",
		txPath,
	}
	for flag, value := range opts {
		onomyCommand = append(onomyCommand, fmt.Sprintf("--%s=%v", flag, value))
	}

	var encoded string
	s.executeOnomyTxCommand(ctx, c, onomyCommand, 0, func(stdOut []byte, stdErr []byte) bool {
		if stdErr != nil {
			return false
		}
		encoded = strings.TrimSuffix(string(stdOut), "\n")
		return true
	})
	s.T().Logf("successfully encode with %v", txPath)
	return encoded
}

func (s *IntegrationTestSuite) execDecode(
	c *chain,
	txPath string,
	opt ...flagOption,
) string {
	opts := applyOptions(c.id, opt)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	s.T().Logf("%s - Executing onomyd decoding with %v", c.id, txPath)
	onomyCommand := []string{
		onomydBinary,
		txCommand,
		"decode",
		txPath,
	}
	for flag, value := range opts {
		onomyCommand = append(onomyCommand, fmt.Sprintf("--%s=%v", flag, value))
	}

	var decoded string
	s.executeOnomyTxCommand(ctx, c, onomyCommand, 0, func(stdOut []byte, stdErr []byte) bool {
		if stdErr != nil {
			return false
		}
		decoded = strings.TrimSuffix(string(stdOut), "\n")
		return true
	})
	s.T().Logf("successfully decode %v", txPath)
	return decoded
}

func (s *IntegrationTestSuite) execVestingTx( //nolint:unused

	c *chain,
	method string,
	args []string,
	opt ...flagOption,
) {
	opts := applyOptions(c.id, opt)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	s.T().Logf("%s - Executing onomyd %s with %v", c.id, method, args)
	onomyCommand := []string{
		onomydBinary,
		txCommand,
		vestingtypes.ModuleName,
		method,
		"-y",
	}
	onomyCommand = append(onomyCommand, args...)

	for flag, value := range opts {
		onomyCommand = append(onomyCommand, fmt.Sprintf("--%s=%v", flag, value))
	}

	s.executeOnomyTxCommand(ctx, c, onomyCommand, 0, s.defaultExecValidation(c, 0))
	s.T().Logf("successfully %s with %v", method, args)
}

func (s *IntegrationTestSuite) execCreatePeriodicVestingAccount( //nolint:unused

	c *chain,
	address,
	jsonPath string,
	opt ...flagOption,
) {
	s.T().Logf("Executing onomyd create periodic vesting account %s", c.id)
	s.execVestingTx(c, "create-periodic-vesting-account", []string{address, jsonPath}, opt...)
	s.T().Logf("successfully created periodic vesting account %s with %s", address, jsonPath)
}

func (s *IntegrationTestSuite) execUnjail(
	c *chain,
	opt ...flagOption,
) {
	opts := applyOptions(c.id, opt)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	s.T().Logf("Executing onomyd slashing unjail %s with options: %v", c.id, opt)
	onomyCommand := []string{
		onomydBinary,
		txCommand,
		slashingtypes.ModuleName,
		"unjail",
		"-y",
	}

	for flag, value := range opts {
		onomyCommand = append(onomyCommand, fmt.Sprintf("--%s=%v", flag, value))
	}

	s.executeOnomyTxCommand(ctx, c, onomyCommand, 0, s.defaultExecValidation(c, 0))
	s.T().Logf("successfully unjail with options %v", opt)
}

func (s *IntegrationTestSuite) execFeeGrant(c *chain, valIdx int, granter, grantee, spendLimit string, opt ...flagOption) {
	opt = append(opt, withKeyValue(flagFrom, granter))
	opt = append(opt, withKeyValue(flagSpendLimit, spendLimit))
	opts := applyOptions(c.id, opt)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	s.T().Logf("granting %s fee from %s on chain %s", grantee, granter, c.id)

	onomyCommand := []string{
		onomydBinary,
		txCommand,
		feegrant.ModuleName,
		"grant",
		granter,
		grantee,
		"-y",
	}
	for flag, value := range opts {
		onomyCommand = append(onomyCommand, fmt.Sprintf("--%s=%v", flag, value))
	}

	s.executeOnomyTxCommand(ctx, c, onomyCommand, valIdx, s.defaultExecValidation(c, valIdx))
}

func (s *IntegrationTestSuite) execFeeGrantRevoke(c *chain, valIdx int, granter, grantee string, opt ...flagOption) {
	opt = append(opt, withKeyValue(flagFrom, granter))
	opts := applyOptions(c.id, opt)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	s.T().Logf("revoking %s fee grant from %s on chain %s", grantee, granter, c.id)

	onomyCommand := []string{
		onomydBinary,
		txCommand,
		feegrant.ModuleName,
		"revoke",
		granter,
		grantee,
		"-y",
	}
	for flag, value := range opts {
		onomyCommand = append(onomyCommand, fmt.Sprintf("--%s=%v", flag, value))
	}

	s.executeOnomyTxCommand(ctx, c, onomyCommand, valIdx, s.defaultExecValidation(c, valIdx))
}

func (s *IntegrationTestSuite) execBankSend(
	c *chain,
	valIdx int,
	from,
	to,
	amt,
	fees string,
	expectErr bool,
	opt ...flagOption,
) {
	// TODO remove the hardcode opt after refactor, all methods should accept custom flags
	opt = append(opt, withKeyValue(flagFees, fees))
	opt = append(opt, withKeyValue(flagFrom, from))
	opts := applyOptions(c.id, opt)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	s.T().Logf("sending %s tokens from %s to %s on chain %s", amt, from, to, c.id)

	onomyCommand := []string{
		onomydBinary,
		txCommand,
		banktypes.ModuleName,
		"send",
		from,
		to,
		amt,
		"-y",
	}
	for flag, value := range opts {
		onomyCommand = append(onomyCommand, fmt.Sprintf("--%s=%v", flag, value))
	}

	s.executeOnomyTxCommand(ctx, c, onomyCommand, valIdx, s.expectErrExecValidation(c, valIdx, expectErr))
}

func (s *IntegrationTestSuite) execBankMultiSend(
	c *chain,
	valIdx int,
	from string,
	to []string,
	amt string,
	fees string,
	expectErr bool,
	opt ...flagOption,
) {
	// TODO remove the hardcode opt after refactor, all methods should accept custom flags
	opt = append(opt, withKeyValue(flagFees, fees))
	opt = append(opt, withKeyValue(flagFrom, from))
	opts := applyOptions(c.id, opt)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	s.T().Logf("sending %s tokens from %s to %s on chain %s", amt, from, to, c.id)

	onomyCommand := []string{
		onomydBinary,
		txCommand,
		banktypes.ModuleName,
		"multi-send",
		from,
	}

	onomyCommand = append(onomyCommand, to...)
	onomyCommand = append(onomyCommand, amt, "-y")

	for flag, value := range opts {
		onomyCommand = append(onomyCommand, fmt.Sprintf("--%s=%v", flag, value))
	}

	s.executeOnomyTxCommand(ctx, c, onomyCommand, valIdx, s.expectErrExecValidation(c, valIdx, expectErr))
}

type txBankSend struct {
	from      string
	to        string
	amt       string
	fees      string
	log       string
	expectErr bool
}

func (s *IntegrationTestSuite) execBankSendBatch(
	c *chain,
	valIdx int, //nolint:unparam
	txs ...txBankSend,
) int {
	sucessBankSendCount := 0

	for i := range txs {
		s.T().Logf(txs[i].log)

		s.execBankSend(c, valIdx, txs[i].from, txs[i].to, txs[i].amt, txs[i].fees, txs[i].expectErr)
		if !txs[i].expectErr {
			if !txs[i].expectErr {
				sucessBankSendCount++
			}
		}
	}

	return sucessBankSendCount
}

func (s *IntegrationTestSuite) execWithdrawAllRewards(c *chain, valIdx int, payee, fees string, expectErr bool) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	onomyCommand := []string{
		onomydBinary,
		txCommand,
		distributiontypes.ModuleName,
		"withdraw-all-rewards",
		fmt.Sprintf("--%s=%s", flags.FlagFrom, payee),
		fmt.Sprintf("--%s=%s", flags.FlagGasPrices, fees),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, c.id),
		"--keyring-backend=test",
		"--output=json",
		"-y",
	}

	s.executeOnomyTxCommand(ctx, c, onomyCommand, valIdx, s.expectErrExecValidation(c, valIdx, expectErr))
}

func (s *IntegrationTestSuite) execDistributionFundCommunityPool(c *chain, valIdx int, from, amt, fees string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	s.T().Logf("Executing onomyd tx distribution fund-community-pool on chain %s", c.id)

	onomyCommand := []string{
		onomydBinary,
		txCommand,
		distributiontypes.ModuleName,
		"fund-community-pool",
		amt,
		fmt.Sprintf("--%s=%s", flags.FlagFrom, from),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, c.id),
		fmt.Sprintf("--%s=%s", flags.FlagFees, fees),
		"--keyring-backend=test",
		"--output=json",
		"-y",
	}

	s.executeOnomyTxCommand(ctx, c, onomyCommand, valIdx, s.defaultExecValidation(c, valIdx))
	s.T().Logf("Successfully funded community pool")
}

func (s *IntegrationTestSuite) runGovExec(c *chain, valIdx int, submitterAddr, govCommand string, proposalFlags []string, fees string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	onomyCommand := []string{
		onomydBinary,
		txCommand,
		govtypes.ModuleName,
		govCommand,
	}

	generalFlags := []string{
		fmt.Sprintf("--%s=%s", flags.FlagFrom, submitterAddr),
		fmt.Sprintf("--%s=%s", flags.FlagGas, "300000"), // default 200000 isn't enough
		fmt.Sprintf("--%s=%s", flags.FlagGasPrices, fees),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, c.id),
		"--keyring-backend=test",
		"--output=json",
		"-y",
	}

	onomyCommand = concatFlags(onomyCommand, proposalFlags, generalFlags)
	s.T().Logf("Executing onomyd tx gov %s on chain %s", govCommand, c.id)
	s.executeOnomyTxCommand(ctx, c, onomyCommand, valIdx, s.defaultExecValidation(c, valIdx))
	s.T().Logf("Successfully executed %s", govCommand)
}

// NOTE: Tx unused, left here for future reference
// func (s *IntegrationTestSuite) executeGKeysAddCommand(c *chain, valIdx int, name string, home string) string {
// 	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
// 	defer cancel()

// 	onomyCommand := []string{
// 		onomydBinary,
// 		keysCommand,
// 		"add",
// 		name,
// 		fmt.Sprintf("--%s=%s", flags.FlagHome, home),
// 		"--keyring-backend=test",
// 		"--output=json",
// 	}

// 	var addrRecord AddressResponse
// 	s.executeOnomyTxCommand(ctx, c, onomyCommand, valIdx, func(stdOut []byte, stdErr []byte) bool {
// 		// Onomyd keys add by default returns payload to stdErr
// 		if err := json.Unmarshal(stdErr, &addrRecord); err != nil {
// 			return false
// 		}
// 		return strings.Contains(addrRecord.Address, "cosmos")
// 	})
// 	return addrRecord.Address
// }

// NOTE: Tx unused, left here for future reference
// func (s *IntegrationTestSuite) executeKeysList(c *chain, valIdx int, home string) { // nolint:U1000
// 	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
// 	defer cancel()

// 	onomyCommand := []string{
// 		onomydBinary,
// 		keysCommand,
// 		"list",
// 		"--keyring-backend=test",
// 		fmt.Sprintf("--%s=%s", flags.FlagHome, home),
// 		"--output=json",
// 	}

// 	s.executeOnomyTxCommand(ctx, c, onomyCommand, valIdx, func([]byte, []byte) bool {
// 		return true
// 	})
// }

func (s *IntegrationTestSuite) execDelegate(c *chain, valIdx int, amount, valOperAddress, delegatorAddr, home, delegateFees string) { //nolint:unparam

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	s.T().Logf("Executing onomyd tx staking delegate %s", c.id)

	onomyCommand := []string{
		onomydBinary,
		txCommand,
		stakingtypes.ModuleName,
		"delegate",
		valOperAddress,
		amount,
		fmt.Sprintf("--%s=%s", flags.FlagFrom, delegatorAddr),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, c.id),
		fmt.Sprintf("--%s=%s", flags.FlagGasPrices, delegateFees),
		"--keyring-backend=test",
		fmt.Sprintf("--%s=%s", flags.FlagHome, home),
		"--output=json",
		"-y",
	}

	s.executeOnomyTxCommand(ctx, c, onomyCommand, valIdx, s.defaultExecValidation(c, valIdx))
	s.T().Logf("%s successfully delegated %s to %s", delegatorAddr, amount, valOperAddress)
}

func (s *IntegrationTestSuite) execUnbondDelegation(c *chain, valIdx int, amount, valOperAddress, delegatorAddr, home, delegateFees string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	s.T().Logf("Executing onomyd tx staking unbond %s", c.id)

	onomyCommand := []string{
		onomydBinary,
		txCommand,
		stakingtypes.ModuleName,
		"unbond",
		valOperAddress,
		amount,
		fmt.Sprintf("--%s=%s", flags.FlagFrom, delegatorAddr),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, c.id),
		fmt.Sprintf("--%s=%s", flags.FlagGasPrices, delegateFees),
		"--keyring-backend=test",
		fmt.Sprintf("--%s=%s", flags.FlagHome, home),
		"--output=json",
		"-y",
	}

	s.executeOnomyTxCommand(ctx, c, onomyCommand, valIdx, s.defaultExecValidation(c, valIdx))
	s.T().Logf("%s successfully undelegated %s to %s", delegatorAddr, amount, valOperAddress)
}

func (s *IntegrationTestSuite) execCancelUnbondingDelegation(c *chain, valIdx int, amount, valOperAddress, creationHeight, delegatorAddr, home, delegateFees string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	s.T().Logf("Executing onomyd tx staking cancel-unbond %s", c.id)

	onomyCommand := []string{
		onomydBinary,
		txCommand,
		stakingtypes.ModuleName,
		"cancel-unbond",
		valOperAddress,
		amount,
		creationHeight,
		fmt.Sprintf("--%s=%s", flags.FlagFrom, delegatorAddr),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, c.id),
		fmt.Sprintf("--%s=%s", flags.FlagGasPrices, delegateFees),
		"--keyring-backend=test",
		fmt.Sprintf("--%s=%s", flags.FlagHome, home),
		"--output=json",
		"-y",
	}

	s.executeOnomyTxCommand(ctx, c, onomyCommand, valIdx, s.defaultExecValidation(c, valIdx))
	s.T().Logf("%s successfully canceled unbonding %s to %s", delegatorAddr, amount, valOperAddress)
}

func (s *IntegrationTestSuite) execRedelegate(c *chain, valIdx int, amount, originalValOperAddress,
	newValOperAddress, delegatorAddr, home, delegateFees string,
) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	s.T().Logf("Executing onomyd tx staking redelegate %s", c.id)

	onomyCommand := []string{
		onomydBinary,
		txCommand,
		stakingtypes.ModuleName,
		"redelegate",
		originalValOperAddress,
		newValOperAddress,
		amount,
		fmt.Sprintf("--%s=%s", flags.FlagFrom, delegatorAddr),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, c.id),
		fmt.Sprintf("--%s=%s", flags.FlagGas, "300000"), // default 200000 isn't enough
		fmt.Sprintf("--%s=%s", flags.FlagGasPrices, delegateFees),
		"--keyring-backend=test",
		fmt.Sprintf("--%s=%s", flags.FlagHome, home),
		"--output=json",
		"-y",
	}

	s.executeOnomyTxCommand(ctx, c, onomyCommand, valIdx, s.defaultExecValidation(c, valIdx))
	s.T().Logf("%s successfully redelegated %s from %s to %s", delegatorAddr, amount, originalValOperAddress, newValOperAddress)
}

func (s *IntegrationTestSuite) getLatestBlockHeight(c *chain, valIdx int) int {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	type syncInfo struct {
		SyncInfo struct {
			LatestHeight string `json:"latest_block_height"`
		} `json:"SyncInfo"`
	}

	var currentHeight int
	onomyCommand := []string{onomydBinary, "status"}
	s.executeOnomyTxCommand(ctx, c, onomyCommand, valIdx, func(stdOut []byte, stdErr []byte) bool {
		var (
			err   error
			block syncInfo
		)
		s.Require().NoError(json.Unmarshal(stdOut, &block))
		currentHeight, err = strconv.Atoi(block.SyncInfo.LatestHeight)
		s.Require().NoError(err)
		return currentHeight > 0
	})
	return currentHeight
}

// func (s *IntegrationTestSuite) verifyBalanceChange(endpoint string, expectedAmount sdk.Coin, recipientAddress string) {
// 	s.Require().Eventually(
// 		func() bool {
// 			afterAtomBalance, err := getSpecificBalance(endpoint, recipientAddress, uatomDenom)
// 			s.Require().NoError(err)

// 			return afterAtomBalance.IsEqual(expectedAmount)
// 		},
// 		20*time.Second,
// 		5*time.Second,
// 	)
// }

func (s *IntegrationTestSuite) execSetWithdrawAddress(
	c *chain,
	valIdx int,
	fees,
	delegatorAddress,
	newWithdrawalAddress,
	homePath string,
) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	s.T().Logf("Setting distribution withdrawal address on chain %s for %s to %s", c.id, delegatorAddress, newWithdrawalAddress)
	onomyCommand := []string{
		onomydBinary,
		txCommand,
		distributiontypes.ModuleName,
		"set-withdraw-addr",
		newWithdrawalAddress,
		fmt.Sprintf("--%s=%s", flags.FlagFrom, delegatorAddress),
		fmt.Sprintf("--%s=%s", flags.FlagFees, fees),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, c.id),
		fmt.Sprintf("--%s=%s", flags.FlagHome, homePath),
		"--keyring-backend=test",
		"--output=json",
		"-y",
	}

	s.executeOnomyTxCommand(ctx, c, onomyCommand, valIdx, s.defaultExecValidation(c, valIdx))
	s.T().Logf("Successfully set new distribution withdrawal address for %s to %s", delegatorAddress, newWithdrawalAddress)
}

func (s *IntegrationTestSuite) execWithdrawReward(
	c *chain,
	valIdx int,
	delegatorAddress,
	validatorAddress,
	homePath string,
) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	s.T().Logf("Withdrawing distribution rewards on chain %s for delegator %s from %s validator", c.id, delegatorAddress, validatorAddress)
	onomyCommand := []string{
		onomydBinary,
		txCommand,
		distributiontypes.ModuleName,
		"withdraw-rewards",
		validatorAddress,
		fmt.Sprintf("--%s=%s", flags.FlagFrom, delegatorAddress),
		fmt.Sprintf("--%s=%s", flags.FlagGasPrices, "300uatom"),
		fmt.Sprintf("--%s=%s", flags.FlagGas, "auto"),
		fmt.Sprintf("--%s=%s", flags.FlagGasAdjustment, "1.5"),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, c.id),
		fmt.Sprintf("--%s=%s", flags.FlagHome, homePath),
		"--keyring-backend=test",
		"--output=json",
		"-y",
	}

	s.executeOnomyTxCommand(ctx, c, onomyCommand, valIdx, s.defaultExecValidation(c, valIdx))
	s.T().Logf("Successfully withdrew distribution rewards for delegator %s from validator %s", delegatorAddress, validatorAddress)
}

func (s *IntegrationTestSuite) executeOnomyTxCommand(ctx context.Context, c *chain, onomyCommand []string, valIdx int, validation func([]byte, []byte) bool) {
	if validation == nil {
		validation = s.defaultExecValidation(s.chainA, 0)
	}
	var (
		outBuf bytes.Buffer
		errBuf bytes.Buffer
	)
	exec, err := s.dkrPool.Client.CreateExec(docker.CreateExecOptions{
		Context:      ctx,
		AttachStdout: true,
		AttachStderr: true,
		Container:    s.valResources[c.id][valIdx].Container.ID,
		User:         "nonroot",
		Cmd:          onomyCommand,
	})
	s.Require().NoError(err)

	err = s.dkrPool.Client.StartExec(exec.ID, docker.StartExecOptions{
		Context:      ctx,
		Detach:       false,
		OutputStream: &outBuf,
		ErrorStream:  &errBuf,
	})
	s.Require().NoError(err)

	stdOut := outBuf.Bytes()
	stdErr := errBuf.Bytes()
	if !validation(stdOut, stdErr) {
		s.Require().FailNowf("Exec validation failed", "stdout: %s, stderr: %s",
			string(stdOut), string(stdErr))
	}
}

func (s *IntegrationTestSuite) executeHermesCommand(ctx context.Context, hermesCmd []string) ([]byte, error) {
	var outBuf bytes.Buffer
	exec, err := s.dkrPool.Client.CreateExec(docker.CreateExecOptions{
		Context:      ctx,
		AttachStdout: true,
		AttachStderr: true,
		Container:    s.hermesResource.Container.ID,
		User:         "root",
		Cmd:          hermesCmd,
	})
	s.Require().NoError(err)

	err = s.dkrPool.Client.StartExec(exec.ID, docker.StartExecOptions{
		Context:      ctx,
		Detach:       false,
		OutputStream: &outBuf,
	})
	s.Require().NoError(err)

	// Check that the stdout output contains the expected status
	// and look for errors, e.g "insufficient fees"
	stdOut := []byte{}
	scanner := bufio.NewScanner(&outBuf)
	for scanner.Scan() {
		stdOut = scanner.Bytes()
		var out map[string]interface{}
		err = json.Unmarshal(stdOut, &out)
		s.Require().NoError(err)
		if err != nil {
			return nil, fmt.Errorf("hermes relayer command returned failed with error: %s", err)
		}
		// errors are catched by observing the logs level in the stderr output
		if lvl := out["level"]; lvl != nil && strings.ToLower(lvl.(string)) == "error" {
			errMsg := out["fields"].(map[string]interface{})["message"]
			return nil, fmt.Errorf("hermes relayer command failed: %s", errMsg)
		}
		if s := out["status"]; s != nil && s != "success" {
			return nil, fmt.Errorf("hermes relayer command returned failed with status: %s", s)
		}
	}

	return stdOut, nil
}

func (s *IntegrationTestSuite) expectErrExecValidation(chain *chain, valIdx int, expectErr bool) func([]byte, []byte) bool {
	return func(stdOut []byte, stdErr []byte) bool {
		var txResp sdk.TxResponse
		gotErr := cdc.UnmarshalJSON(stdOut, &txResp) != nil
		if gotErr {
			s.Require().True(expectErr)
		}

		endpoint := fmt.Sprintf("http://%s", s.valResources[chain.id][valIdx].GetHostPort("1317/tcp"))
		// wait for the tx to be committed on chain
		s.Require().Eventuallyf(
			func() bool {
				gotErr := queryOnomyTx(endpoint, txResp.TxHash) != nil
				return gotErr == expectErr
			},
			time.Minute,
			5*time.Second,
			"stdOut: %s, stdErr: %s",
			string(stdOut), string(stdErr),
		)
		return true
	}
}

func (s *IntegrationTestSuite) defaultExecValidation(chain *chain, valIdx int) func([]byte, []byte) bool {
	return func(stdOut []byte, stdErr []byte) bool {
		var txResp sdk.TxResponse
		if err := cdc.UnmarshalJSON(stdOut, &txResp); err != nil {
			return false
		}
		if strings.Contains(txResp.String(), "code: 0") || txResp.Code == 0 {
			endpoint := fmt.Sprintf("http://%s", s.valResources[chain.id][valIdx].GetHostPort("1317/tcp"))
			s.Require().Eventually(
				func() bool {
					return queryOnomyTx(endpoint, txResp.TxHash) == nil
				},
				time.Minute,
				5*time.Second,
				"stdOut: %s, stdErr: %s",
				string(stdOut), string(stdErr),
			)
			return true
		}
		return false
	}
}

func (s *IntegrationTestSuite) executeValidatorBond(c *chain, valIdx int, valOperAddress, delegatorAddr, home, delegateFees string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	s.T().Logf("Executing onomyd tx staking validator-bond %s", c.id)

	onomyCommand := []string{
		onomydBinary,
		txCommand,
		stakingtypes.ModuleName,
		"validator-bond",
		valOperAddress,
		fmt.Sprintf("--%s=%s", flags.FlagFrom, delegatorAddr),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, c.id),
		fmt.Sprintf("--%s=%s", flags.FlagGasPrices, delegateFees),
		"--keyring-backend=test",
		fmt.Sprintf("--%s=%s", flags.FlagHome, home),
		"--output=json",
		"-y",
	}

	s.executeOnomyTxCommand(ctx, c, onomyCommand, valIdx, s.defaultExecValidation(c, valIdx))
	s.T().Logf("%s successfully executed validator bond tx to %s", delegatorAddr, valOperAddress)
}

func (s *IntegrationTestSuite) executeTokenizeShares(c *chain, valIdx int, amount, valOperAddress, delegatorAddr, home, delegateFees string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	s.T().Logf("Executing onomyd tx staking tokenize-share %s", c.id)

	onomyCommand := []string{
		onomydBinary,
		txCommand,
		stakingtypes.ModuleName,
		"tokenize-share",
		valOperAddress,
		amount,
		delegatorAddr,
		fmt.Sprintf("--%s=%s", flags.FlagFrom, delegatorAddr),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, c.id),
		fmt.Sprintf("--%s=%s", flags.FlagGasPrices, delegateFees),
		fmt.Sprintf("--%s=%d", flags.FlagGas, 1000000),
		"--keyring-backend=test",
		fmt.Sprintf("--%s=%s", flags.FlagHome, home),
		"--output=json",
		"-y",
	}

	s.executeOnomyTxCommand(ctx, c, onomyCommand, valIdx, s.defaultExecValidation(c, valIdx))
	s.T().Logf("%s successfully executed tokenize share tx from %s", delegatorAddr, valOperAddress)
}

func (s *IntegrationTestSuite) executeRedeemShares(c *chain, valIdx int, amount, delegatorAddr, home, delegateFees string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	s.T().Logf("Executing onomyd tx staking redeem-tokens %s", c.id)

	onomyCommand := []string{
		onomydBinary,
		txCommand,
		stakingtypes.ModuleName,
		"redeem-tokens",
		amount,
		fmt.Sprintf("--%s=%s", flags.FlagFrom, delegatorAddr),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, c.id),
		fmt.Sprintf("--%s=%s", flags.FlagGasPrices, delegateFees),
		fmt.Sprintf("--%s=%d", flags.FlagGas, 1000000),
		"--keyring-backend=test",
		fmt.Sprintf("--%s=%s", flags.FlagHome, home),
		"--output=json",
		"-y",
	}

	s.executeOnomyTxCommand(ctx, c, onomyCommand, valIdx, s.defaultExecValidation(c, valIdx))
	s.T().Logf("%s successfully executed redeem share tx for %s", delegatorAddr, amount)
}

func (s *IntegrationTestSuite) executeTransferTokenizeShareRecord(c *chain, valIdx int, recordID, owner, newOwner, home, txFees string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	s.T().Logf("Executing onomyd tx staking transfer-tokenize-share-record %s", c.id)

	onomyCommand := []string{
		onomydBinary,
		txCommand,
		stakingtypes.ModuleName,
		"transfer-tokenize-share-record",
		recordID,
		newOwner,
		fmt.Sprintf("--%s=%s", flags.FlagFrom, owner),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, c.id),
		fmt.Sprintf("--%s=%s", flags.FlagGasPrices, txFees),
		"--keyring-backend=test",
		fmt.Sprintf("--%s=%s", flags.FlagHome, home),
		"--output=json",
		"-y",
	}

	s.executeOnomyTxCommand(ctx, c, onomyCommand, valIdx, s.defaultExecValidation(c, valIdx))
	s.T().Logf("%s successfully executed transfer tokenize share record for %s", owner, recordID)
}

// signTxFileOnline signs a transaction file using the onomycli tx sign command
// the from flag is used to specify the keyring account to sign the transaction
// the from account must be registered in the keyring and exist on chain (have a balance or be a genesis account)
func (s *IntegrationTestSuite) signTxFileOnline(chain *chain, valIdx int, from string, txFilePath string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	onomyCommand := []string{
		onomydBinary,
		txCommand,
		"sign",
		filepath.Join(onomyHomePath, txFilePath),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, chain.id),
		fmt.Sprintf("--%s=%s", flags.FlagHome, onomyHomePath),
		fmt.Sprintf("--%s=%s", flags.FlagFrom, from),
		"--keyring-backend=test",
		"--output=json",
		"-y",
	}

	var output []byte
	var erroutput []byte
	captureOutput := func(stdout []byte, stderr []byte) bool {
		output = stdout
		erroutput = stderr
		return true
	}

	s.executeOnomyTxCommand(ctx, chain, onomyCommand, valIdx, captureOutput)
	if len(erroutput) > 0 {
		return nil, fmt.Errorf("failed to sign tx: %s", string(erroutput))
	}
	return output, nil
}

// broadcastTxFile broadcasts a signed transaction file using the onomycli tx broadcast command
// the from flag is used to specify the keyring account to sign the transaction
// the from account must be registered in the keyring and exist on chain (have a balance or be a genesis account)
func (s *IntegrationTestSuite) broadcastTxFile(chain *chain, valIdx int, from string, txFilePath string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	broadcastTxCmd := []string{
		onomydBinary,
		txCommand,
		"broadcast",
		filepath.Join(onomyHomePath, txFilePath),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, chain.id),
		fmt.Sprintf("--%s=%s", flags.FlagHome, onomyHomePath),
		fmt.Sprintf("--%s=%s", flags.FlagFrom, from),
		"--keyring-backend=test",
		"--output=json",
		"-y",
	}

	var output []byte
	var erroutput []byte
	captureOutput := func(stdout []byte, stderr []byte) bool {
		output = stdout
		erroutput = stderr
		return true
	}

	s.executeOnomyTxCommand(ctx, chain, broadcastTxCmd, valIdx, captureOutput)
	if len(erroutput) > 0 {
		return nil, fmt.Errorf("failed to sign tx: %s", string(erroutput))
	}
	return output, nil
}