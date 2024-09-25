package proxy

import (
	"io"
	"log"
	"net"
	"salmonproxy/hosts"
	"salmonproxy/mc"
)

// Proxy - Manages a Proxy connection, piping data between local and remote.
type Proxy struct {
	sentBytes     uint64
	receivedBytes uint64
	laddr         *net.TCPAddr
	lconn         *net.TCPConn
	errSig        chan bool
}

// New - Create a new Proxy instance. Takes over local connection passed in,
// and closes it when finished.
func New(lconn *net.TCPConn, laddr *net.TCPAddr) *Proxy {
	return &Proxy{
		lconn:  lconn,
		laddr:  laddr,
		errSig: make(chan bool),
	}
}

// Start - open connection to remote and start proxying data.
func (p *Proxy) Start() {
	defer p.lconn.Close()

	var err error

	log.Println("Accepted new connection on:", p.laddr, "from:", p.lconn.RemoteAddr())
	log.Println("Reading minecraft data")
	fp := mc.FirstPacket{}
	buffer, err := mc.ReadFirstPacket(p.lconn, &fp)
	if err != nil {
		log.Printf("Error reading minecraft data: %s\n", err)
		return
	}

	host, err := hosts.GetHost(fp.ServerAddress)
	if err != nil {
		log.Println(err)
		return
	}
	raddr, err := net.ResolveTCPAddr("tcp", host)
	if err != nil {
		log.Printf("Failed to resolve remote address: %s\n", err)
		return
	}

	//connect to remote
	rconn, err := net.DialTCP("tcp", nil, raddr)
	if err != nil {
		log.Printf("Remote connection failed: %s\n", err)
		return
	}
	defer rconn.Close()

	// write first packet
	rconn.Write(buffer)

	//display both ends
	log.Printf("Opened %s >>> %s(%s)\n", p.lconn.RemoteAddr(), fp.ServerAddress, raddr.String())

	//bidirectional copy
	go pipe(p.lconn, rconn, p.errSig, &p.sentBytes)
	go pipe(rconn, p.lconn, p.errSig, &p.receivedBytes)

	//wait for close...
	<-p.errSig
	log.Printf("Closed %s >>> %s(%s)  (%d bytes sent, %d bytes recieved)\n", p.lconn.RemoteAddr(), fp.ServerAddress, raddr.String(), p.sentBytes, p.receivedBytes)
}

func pipe(src, dst io.ReadWriter, errSig chan bool, bytesCounter *uint64) {
	//directional copy (64k buffer)
	buff := make([]byte, 0xffff)
	for {
		bytesRead, err := src.Read(buff)
		if err != nil {
			log.Printf("Read failed '%s'\n", err)
			errSig <- true
		}
		b := buff[:bytesRead]

		//write out result
		bytesWritten, err := dst.Write(b)
		if err != nil {
			log.Printf("Write failed '%s'\n", err)
			errSig <- true
		}
		*bytesCounter += uint64(bytesWritten)
	}
}
