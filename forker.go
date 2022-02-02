package main
import (
	"fmt"
	"time"
	"strings"
	"strconv"
	"math/rand"
)
var Number int
var ForkersNumber int
const (
	NodeNumber     = 20
	MaxChannelSize = 100
	ratio = 0.3
)

type BlockChain struct {
	chain []Block
}

type Block struct {
    PoW int64
	hash int64
    PreviousHash int64
}

// May need other fields
type Node struct {
	id          uint64
	peers       map[uint64]chan Message
	receiveChan chan Message
	chain 		*BlockChain
}

// Define your message's struct here
type Message struct {
	sender uint64
	blockchain 	   BlockChain
}

func Random() int64{
	var test strings.Builder
	rand.Seed(time.Now().UnixNano())
	for i := 1; i < 7; i++ {
		RandInt := rand.Intn(9) + 1
		test.WriteString(strconv.Itoa(RandInt))
	}
	s,err := strconv.Atoi(test.String())
	if err == nil {
		return int64(s)
	} else {
        return int64(0)
    }
}

func (blockchain *BlockChain) NewBlock(pow int64, previousHash int64) Block {
	ph := previousHash
    newBlock := Block{
        PoW: pow,
		hash : Random(),
        PreviousHash: ph,
    }
    blockchain.chain = append(blockchain.chain, newBlock)
    return newBlock
}

func (blockchain *BlockChain) FirstBlock(pow int64, previousHash int64) Block {
	ph := previousHash
    newBlock := Block{
        PoW: pow,
		hash : Random(),
        PreviousHash: ph,
    }
    blockchain.chain = append(blockchain.chain, newBlock)
    return newBlock
}

func (blockchain *BlockChain) ProofOfWork(lastProof int64) int64 {
	num := 1000 + (rand.Intn(9) + 1)*50
    time.Sleep(time.Duration(num) * time.Millisecond)
    return Random()
}

func (blockchain *BlockChain) IsChainOk(chain *[]Block) bool {
    lastBlock := (*chain)[0]
    currentIndex := 1
    for currentIndex < len(*chain) {
        block := (*chain)[currentIndex]
        // Check that the hash of the block is correct
        if block.PreviousHash != lastBlock.hash {
            return false
        }
        lastBlock = block
        currentIndex += 1
    }
    return true
}

func NewBlockChain() *BlockChain {
    newBlockChain := &BlockChain{
        chain:        make([]Block, 0),
    }
    // Initial, sentinel block
    newBlockChain.FirstBlock(111, 0000)
    return newBlockChain
}

func NewNode(id uint64, peers map[uint64]chan Message, recvChan chan Message, blockchain *BlockChain) *Node {
	return &Node{
		id:          id,
		peers:       peers,
		receiveChan: recvChan,
		chain:		 blockchain,
	}
}

func (n *Node) Run() {
	fmt.Println("start node : ", n.id)
	go n.Receive()
	go n.NormalMining()
}

func (n *Node) ForkerRun() {
	fmt.Println("start attack node : ", n.id)
	go n.Receive()
	go n.ForkerMining()
}

func (n *Node) NormalMining() {
	for {
		blockchain := n.chain
		proof := blockchain.ProofOfWork(blockchain.chain[len(blockchain.chain)-1].PoW)
		blockchain.NewBlock(proof, blockchain.chain[len(blockchain.chain)-1].hash)		
		if (blockchain.IsChainOk(&(blockchain.chain))) {
			Number = len(n.chain.chain)
			n.Broadcast(Message{sender : n.id, blockchain : *(n.chain)})
		}
	}
}

func (n *Node) ForkerMining() {
	for {
		blockchain := n.chain
		proof := blockchain.ProofOfWork(blockchain.chain[len(blockchain.chain)-1].PoW)
		blockchain.NewBlock(proof, blockchain.chain[len(blockchain.chain)-1].hash)
		if (blockchain.IsChainOk(&(blockchain.chain))) {
			ForkersNumber := len(n.chain.chain)
			if (ForkersNumber > Number + 2) {
				fmt.Println("Attack!")
			}
			n.Broadcast(Message{sender : n.id, blockchain : *(n.chain)})
		}
	}
}

func (n *Node) Receive() {
	for {
		select {
		case msg := <-n.receiveChan:
			n.handler(msg)
		}
	}
}

func (n *Node) handler(msg Message) {
	fmt.Println("Node", n.id, "received message from node", msg.sender)
}

func (n *Node) Broadcast(msg Message) {
	for id, ch := range n.peers {
		if id == n.id {
			continue
		}
		ch <- msg
	}
}

func main() {

	nodes := make([]*Node, int(NodeNumber*(1-ratio)))
	peers := make(map[uint64]chan Message)
	for i := 0; i < int(NodeNumber*(1-ratio)); i++ {
		peers[uint64(i)] = make(chan Message, MaxChannelSize)
	}
	for i := uint64(0); i < uint64(NodeNumber*(1-ratio)); i++ {
		var blockchain = NewBlockChain()
		nodes[i] = NewNode(i, peers, peers[i], blockchain)
	}

	forkers := make([]*Node, int(NodeNumber*ratio))
	peersF := make(map[uint64]chan Message)
	for i := 0; i < int(NodeNumber*(ratio)); i++ {
		peersF[uint64(i)] = make(chan Message, MaxChannelSize)
	}
	for i := uint64(0); i < uint64(NodeNumber*(ratio)); i++ {
		var blockchain = NewBlockChain()
		forkers[i] = NewNode(i, peersF, peersF[i], blockchain)
	}

	//start all nodes
	for i := 0; i < int(NodeNumber*(1-ratio)); i++ {
		go nodes[i].Run()
	}
	for i := 0; i < int(NodeNumber*ratio); i++ {
		go forkers[i].ForkerRun()
	}

	// block to wait for all nodes' threads
	<-make(chan int)
}