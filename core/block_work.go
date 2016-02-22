package core

import (
	"github.com/ethereum/go-ethereum/balancer"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/pow"
)

type nonceResult struct {
	index int
	valid bool
}

func BalanceTxWork(b *balancer.Balancer, txs types.Transactions) {
	if len(txs) == 0 {
		return
	}

	errch := make(chan error, len(txs))

	for i := 0; i < len(txs); i++ {
		i := i
		// create new tasks
		task := balancer.NewTask(func() error {
			txs[i].FromFrontier()
			return nil
		}, errch)

		b.Push(task)
	}

	// we aren't at all interested in the errors
	// since we handle errors ourself.
	go func() {
		for i := 0; i < cap(errch); i++ {
			<-errch
		}
		close(errch)
	}()
}

func BalanceBlockWork(b *balancer.Balancer, blocks []*types.Block, checker pow.PoW) chan nonceResult {
	if len(blocks) == 0 {
		return nil
	}

	var (
		nonceResults = make(chan nonceResult, len(blocks))
		errch        = make(chan error, len(blocks))
	)
	for i, block := range blocks {
		i := i
		task := balancer.NewTask(func() error {
			valid := checker.Verify(block)
			for _, u := range block.Uncles() {
				if !checker.Verify(types.NewBlockWithHeader(u)) {
					valid = false
					break
				}
			}
			nonceResults <- nonceResult{i, valid}
			return nil
		}, errch)
		b.Push(task)

		BalanceTxWork(b, block.Transactions())
	}

	// we aren't at all interested in the errors
	// since we handle errors ourself.
	go func() {
		for i := 0; i < cap(errch); i++ {
			<-errch
		}
		close(errch)
	}()
	return nonceResults
}

/*

const taskCount = 20

func BalanceTxWork(b *balancer.Balancer, txs types.Transactions) {
	if len(txs) == 0 {
		return
	}

	workSize := len(txs) / taskCount
	if workSize == 0 {
		workSize = 1
	}

	errch := make(chan error, int(math.Ceil(float64(len(txs))/float64(workSize)))) // error channel (buffered)
	for i := 0; i < len(txs); i += workSize {
		max := int(math.Min(float64(i+workSize), float64(len(txs)))) // get max size...
		batch := txs[i:max]                                          // ...and create batch

		batchNo := i // batch number for task
		// create new tasks
		task := balancer.NewTask(func() error {
			for i := 0; i < max-batchNo; i++ {
				batch[i].FromFrontier()
			}
			return nil
		}, errch)

		b.Push(task)
	}

	// we aren't at all interested in the errors
	// since we handle errors ourself.
	go func() {
		for i := 0; i < cap(errch); i++ {
			<-errch
		}
		close(errch)
	}()
}
func BalanceBlockWork(b *balancer.Balancer, blocks []*types.Block, checker pow.PoW) chan nonceResult {
	workSize := len(blocks) / taskCount
	if workSize == 0 {
		workSize = 1
	}

	var (
		nonceResults = make(chan nonceResult, len(blocks))                                      // the nonce result channel (buffered)
		errch        = make(chan error, int(math.Ceil(float64(len(blocks))/float64(workSize)))) // error channel (buffered)
	)
	for i := 0; i < len(blocks); i += workSize {
		max := int(math.Min(float64(i+workSize), float64(len(blocks)))) // get max size...
		batch := blocks[i:max]                                          // ...and create batch

		batchNo := i // batch number for task
		// create new tasks
		task := balancer.NewTask(func() error {
			for i := 0; i < max-batchNo; i++ {
				valid := checker.Verify(batch[i])
				for _, u := range batch[i].Uncles() {
					if !checker.Verify(types.NewBlockWithHeader(u)) {
						valid = false
						break
					}
				}
				nonceResults <- nonceResult{batchNo + i, valid}
			}
			return nil
		}, errch)

		b.Push(task)
	}

	// we aren't at all interested in the errors
	// since we handle errors ourself.
	go func() {
		for i := 0; i < cap(errch); i++ {
			<-errch
		}
		close(errch)
	}()

	return nonceResults
}
*/
