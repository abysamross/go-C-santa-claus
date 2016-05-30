package main

import (
	"fmt"
	"sync"
	"time"
)

const (
	ID     = 0
	STATUS = 1
)

const (
	STOP      = 0
	WORK      = 1
	DONE      = 2
	READY     = 3
	SERVED    = 4
	NOTSERVED = 5
)

var (
	TOT_REINDEERS    = 9
	TOT_ELVES        = 14
	CRIT_ELVES_COUNT = 3
	NUMBER_OF_RUNS   = 100
)

type Deer struct {
	id      int
	status  int
	service int
}

type Elfie struct {
	id      int
	status  int
	service int
}

func main() {

	var (
		tempE    = 0
		tempRuns = 0
	)

	fmt.Println("\nSanta Claus Problem Solution in Go Language")
	fmt.Println("*******************************************")
	fmt.Printf("Current # of reindeers: %d\n", TOT_REINDEERS)
	fmt.Printf("Current # of elves: %d\n", TOT_ELVES)
	fmt.Printf("Current # of Runs: %d\n", NUMBER_OF_RUNS)
	fmt.Println("NOTE: change value of NUMBER_OF_RUNS constant in code to run the simulation for longer time")
	fmt.Println("*******************************************\n")

	fmt.Println("If you wish to change the number of elves enter below. (Enter 0 if you want to keep the default value)")
	n1, err1 := fmt.Scanf("%d\n", &tempE)

	fmt.Println("If you wish to change the number of runs enter below. (Enter 0 if you want to keep the default value)")
	n2, err2 := fmt.Scanf("%d\n", &tempRuns)

	if err1 != nil || n1 != 1 {
		fmt.Println(n1, err1)
		fmt.Println("Enter a valid number of elves")
	}
	if err2 != nil || n2 != 1 {
		fmt.Println(n2, err2)
		fmt.Println("Enter a valid number of elves")
	}
	if tempE > 0 {
		TOT_ELVES = tempE
	}
	if tempRuns > 0 {
		NUMBER_OF_RUNS = tempRuns
	}
	fmt.Println("*******************************************\n")
	var wg sync.WaitGroup

	sendReindeers_chan := make([]chan Deer, TOT_REINDEERS)
	sendElves_chan := make([]chan Elfie, TOT_ELVES)

	recvReindeers_chan := make(chan Deer)
	recvElves_chan := make(chan Elfie)

	for i := range sendReindeers_chan {
		sendReindeers_chan[i] = make(chan Deer)

		wg.Add(1)
		go func(i int) {
			Reindeer(i, recvReindeers_chan, sendReindeers_chan[i])
			wg.Done()
		}(i)
	}

	for j := range sendElves_chan {
		sendElves_chan[j] = make(chan Elfie)

		wg.Add(1)
		go func(j int) {
			Elf(j, recvElves_chan, sendElves_chan[j])
			wg.Done()
		}(j)

	}

	wg.Add(1)
	go func() {
		Santa(recvReindeers_chan, sendReindeers_chan, recvElves_chan, sendElves_chan)
		wg.Done()
	}()
	wg.Wait()

	fmt.Println("\n************ END OF MAIN ***************\n")
}

func Santa(recvRs_chan chan Deer, sendRs_chan []chan Deer, recvEs_chan chan Elfie, sendEs_chan []chan Elfie) {
	fmt.Println("\nSanta Ready...\n")

	var rCount int
	var reindeers_chan []chan Deer

	resetReindeers := func() {
		reindeers_chan = make([]chan Deer, TOT_REINDEERS)
		rCount = 0
	}
	resetReindeers()

	var eCount int
	var elves_chan []chan Elfie
	resetElves := func() {
		elves_chan = make([]chan Elfie, CRIT_ELVES_COUNT)
		eCount = 0
	}
	resetElves()

	//Receive from reindeer and elves DONE
	//But they use the sendEs_chan[id] or sendRs_chan[id]
	//itself to send the DONE msg
	wait_deer := func(helperRecv chan Deer) {
		for {
			select {
			case deer := <-helperRecv:
				switch deer.status {
				case DONE:
					return
				default:
				}

			}
		}
	}

	wait_elf := func(helperRecv chan Elfie) {
		for {
			select {
			case elf := <-helperRecv:
				switch elf.status {
				case DONE:
					return
				default:
				}

			}
		}
	}

	for i := 0; i < NUMBER_OF_RUNS; i++ {

		fmt.Printf("									  RUNS #: %-80v\n", i+1)
		select {

		case rArrived := <-recvRs_chan:
			if reindeerStatus := rArrived.status; reindeerStatus == READY {
				fmt.Printf("Deer %d is back from holiday\n", rArrived.id+1)
				reindeers_chan[rArrived.id] = sendRs_chan[rArrived.id]
				rCount++

				if rCount == TOT_REINDEERS {
					fmt.Println("\n===>>> ALL REINDEERS ARE BACK FROM HOLIDAY <<<===\n")
					for _, reindeer := range reindeers_chan {
						reindeer <- Deer{rArrived.id, WORK, SERVED}
						wait_deer(reindeer)
					}
					resetReindeers()
				}
			}

		case eArrived := <-recvEs_chan:
			if elfStatus := eArrived.status; elfStatus == READY {
				fmt.Printf(" Help !!! Elf %d needs some help\n", eArrived.id+1)
				elves_chan[eCount] = sendEs_chan[eArrived.id]
				eCount++

				//Implementing priority: Even if we received
				//something in elves receive chan, before checking
				//if elves count is 3, we check if anything more
				//is coming in reindeer receive and will that take
				//reindeer count to 9
				select {
				case r_e_Arrived := <-recvRs_chan:
					if r_e_Status := r_e_Arrived.status; r_e_Status == READY {
						fmt.Printf("Deer %d is back from holiday\n", r_e_Arrived.id+1)
						reindeers_chan[r_e_Arrived.id] = sendRs_chan[r_e_Arrived.id]
						rCount++

						if rCount == TOT_REINDEERS {
							fmt.Println("\n===>>> ALL REINDEERS ARE BACK FROM HOLIDAY <<<===\n")
							for _, reindeer := range reindeers_chan {
								reindeer <- Deer{r_e_Arrived.id, WORK, SERVED}
								wait_deer(reindeer)
							}
							resetReindeers()
						}
					}
				default:
				}
				if eCount == CRIT_ELVES_COUNT {
					fmt.Println("\n<<<=== SERVICE 3 NEEDY ELVES ===>>>\n")

					for _, elf := range elves_chan {
						elf <- Elfie{eArrived.id, WORK, SERVED}
						wait_elf(elf)
					}
					//fmt.Println("<<<=== DONE SERVICING THE 3 NEEDY ELVES ===>>>")
					resetElves()
				}
			}
		case <-time.After(time.Second * 3):
			fmt.Printf("Timed Out!\n")
			break
		}
		//fmt.Printf("******** END RUN ## %v **************************\n", i+1)
	}

	fmt.Printf("\n***************************** %d RUNS OVER *******************************\n\n", NUMBER_OF_RUNS)
	/*My addition to account for left over unserviced deers*/
	for i, s := range sendRs_chan {
		s <- Deer{i, STOP, SERVED}
	}
	//fmt.Println("============ Stoppped all Reindeers =============")

	/*My addition to account for left over unserviced elves*/
	for i := eCount; i != TOT_ELVES-1; {
		select {
		case remaining_eArrived := <-recvEs_chan:
			if elfStatus := remaining_eArrived.status; elfStatus == READY {
				fmt.Printf(" Help !!! Elf %d needs some help\n", remaining_eArrived.id+1)
				i = remaining_eArrived.id
				elves_chan[eCount] = sendEs_chan[remaining_eArrived.id]
				eCount++

				//if TOT_ELVES i
				if eCount == CRIT_ELVES_COUNT {
					fmt.Println("\n<<<=== SERVICING REMAINING ELVES IN NEEDY GROUPS OF 3 ===>>>\n")

					for _, elf := range elves_chan {
						elf <- Elfie{remaining_eArrived.id, WORK, SERVED}
						wait_elf(elf)
					}
					//fmt.Println("<<<=== DONE SERVICING THE 3 NEEDY ELVES ===>>>")
					resetElves()
				}
			}
		case <-time.After(time.Second * 3):
			fmt.Printf("Timed Out!!!\n")
			break
		}
	}
	//fmt.Println("============ Serviced remaining un-serviced Elves =============")

	fmt.Println("\n============ Stopping all Elves =============\n")
	for i, s := range sendEs_chan {
		s <- Elfie{i, STOP, SERVED}
	}

	for _, s := range sendRs_chan {
		close(s)
	}

	for _, s := range sendEs_chan {
		close(s)
	}

}

func Reindeer(id int, sendSanta chan Deer, recvSanta chan Deer) {

	serviced := false
	for {

		select {
		case sendSanta <- Deer{id, READY, NOTSERVED}:
			//fmt.Printf("Deer %d is back from holiday\n", id)

			select {
			case santaMsg := <-recvSanta:
				serviced = false
				switch santaMsg.status {
				case WORK:
					if santaMsg.service == SERVED {
						fmt.Printf("Deer %d is delivering presents\n", id+1)
						serviced = true
						recvSanta <- Deer{id, DONE, SERVED}
						time.Sleep(500 * time.Millisecond)
					}
				default:
				}
			}
		case santaMsg := <-recvSanta:
			switch santaMsg.status {
			case STOP:
				if serviced == true {
					fmt.Printf("Deer %d terminating, pulled the sleigh. Delivered presents = %v...\n", id+1, serviced)
				} else {
					fmt.Printf("Deer %d getting terminated without pulling the sleigh. Delivered presents = %v...\n", id+1, serviced)
				}
				return
			}
		}
	}
}

func Elf(id int, sendSanta chan Elfie, recvSanta chan Elfie) {

	serviced := false
	for {

		select {
		case sendSanta <- Elfie{id, READY, NOTSERVED}:
			//fmt.Printf("Elf %d needs some help\n", id)

			select {
			case santaMsg := <-recvSanta:
				serviced = false
				switch santaMsg.status {
				case WORK:
					if santaMsg.service == SERVED {
						fmt.Printf(" Phew !!! Elf %d is getting helped by Santa\n", id+1)
						serviced = true
						recvSanta <- Elfie{id, DONE, SERVED}
						time.Sleep(500 * time.Millisecond)
					}
					//default:
				}
			}
		case santaMsg := <-recvSanta:
			switch santaMsg.status {
			case STOP:
				if serviced == true {
					fmt.Printf("Elf %d terminating, got helped by Santa. Serviced = %v...\n", id+1, serviced)
				} else {
					fmt.Printf("Elf %d getting terminated without getting helped by Santa. Serviced =  %v...\n", id+1, serviced)
				}
				return
			}
		}
	}
}
