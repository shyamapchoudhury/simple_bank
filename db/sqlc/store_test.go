package db

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTeansferTx(t *testing.T) {
	store := NewStore(testDB)

	acc1 := createRandomAccount(t)
	acc2 := createRandomAccount(t)
	fmt.Println(">> before:", acc1.Balance, acc2.Balance)

	n := 2
	amount := int64(10)
	errs := make(chan error)
	results := make(chan TransferTxResult)

	for i := 0; i < n; i++ {
		txName := fmt.Sprintf("tx %d", i+1)
		go func() {
			ctx := context.WithValue(context.Background(), txKey, txName)
			resut, err := store.TransferTx(ctx, TransferTxParams{
				FromAccountID: acc1.ID,
				ToAccountID:   acc2.ID,
				Amount:        amount,
			})

			errs <- err
			results <- resut
		}()
	}
	existed := make(map[int]bool)
	for i := 0; i < n; i++ {
		err := <-errs
		require.NoError(t, err)
		result := <-results
		require.NotEmpty(t, result)

		transfer := result.Transfer
		require.NotEmpty(t, transfer)

		require.Equal(t, sql.NullInt64{Int64: acc1.ID, Valid: true}, transfer.FromAccountID)
		require.Equal(t, sql.NullInt64{Int64: acc2.ID, Valid: true}, transfer.ToAccountID)
		require.Equal(t, amount, transfer.Amount)
		require.NotZero(t, transfer.ID)
		require.NotZero(t, transfer.CreatedAt)

		_, err = store.GetTransfer(context.Background(), transfer.ID)
		require.NoError(t, err)

		fromEntry := result.FromEntry
		require.NotEmpty(t, fromEntry)
		require.Equal(t, sql.NullInt64{Int64: acc1.ID, Valid: true}, fromEntry.AccountID)
		require.Equal(t, -amount, fromEntry.Amount)
		require.NotZero(t, fromEntry.ID)
		require.NotZero(t, fromEntry.CreatedAt)

		_, err = store.GetEntry(context.Background(), fromEntry.ID)
		require.NoError(t, err)

		toEntry := result.ToEntry
		require.NotEmpty(t, toEntry)
		require.Equal(t, sql.NullInt64{Int64: acc2.ID, Valid: true}, toEntry.AccountID)
		require.Equal(t, amount, toEntry.Amount)
		require.NotZero(t, toEntry.ID)
		require.NotZero(t, toEntry.CreatedAt)

		_, err = store.GetEntry(context.Background(), toEntry.ID)
		require.NoError(t, err)

		fromAcount := result.FromAccount
		require.NotEmpty(t, fromAcount)
		require.Equal(t, acc1.ID, fromAcount.ID)

		toAcount := result.ToAccount
		require.NotEmpty(t, toAcount)
		require.Equal(t, acc2.ID, toAcount.ID)

		// TODO: check account balance
		fmt.Println(">> Tx:", fromAcount.Balance, toAcount.Balance)
		diff1 := acc1.Balance - fromAcount.Balance
		diff2 := toAcount.Balance - acc2.Balance
		require.Equal(t, diff1, diff2)
		require.True(t, diff1 > 0)
		require.True(t, diff1%amount == 0)

		k := int(diff1 / amount)
		require.True(t, k >= 1 && k <= n)
		require.NotContains(t, existed, k)
		existed[k] = true

	}

	updatedAcc1, err := testQueries.GetAccount(context.Background(), acc1.ID)
	require.NoError(t, err)

	updatedAcc2, err := testQueries.GetAccount(context.Background(), acc2.ID)
	require.NoError(t, err)

	fmt.Println(">> after:", updatedAcc1.Balance, updatedAcc2.Balance)
	require.Equal(t, (acc1.Balance - int64(n)*amount), updatedAcc1.Balance)
	require.Equal(t, (acc2.Balance + int64(n)*amount), updatedAcc2.Balance)

}

func TestTeansferTxDeadLock(t *testing.T) {
	store := NewStore(testDB)

	acc1 := createRandomAccount(t)
	acc2 := createRandomAccount(t)
	fmt.Println(">> before:", acc1.Balance, acc2.Balance)

	n := 10
	amount := int64(10)
	errs := make(chan error)

	for i := 0; i < n; i++ {
		fromAccountId := acc1.ID
		toAccountId := acc2.ID

		if i%2 == 1 {
			fromAccountId = acc2.ID
			toAccountId = acc1.ID
		}
		go func() {

			_, err := store.TransferTx(context.Background(), TransferTxParams{
				FromAccountID: fromAccountId,
				ToAccountID:   toAccountId,
				Amount:        amount,
			})

			errs <- err
		}()
	}

	for i := 0; i < n; i++ {
		err := <-errs
		require.NoError(t, err)

	}

	updatedAcc1, err := testQueries.GetAccount(context.Background(), acc1.ID)
	require.NoError(t, err)

	updatedAcc2, err := testQueries.GetAccount(context.Background(), acc2.ID)
	require.NoError(t, err)

	fmt.Println(">> after:", updatedAcc1.Balance, updatedAcc2.Balance)
	require.Equal(t, acc1.Balance, updatedAcc1.Balance)
	require.Equal(t, acc2.Balance, updatedAcc2.Balance)

}
