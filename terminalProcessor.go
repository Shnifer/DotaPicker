package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

type AnzHero struct {
	name   string
	medAdv float64
	medWRvs float64
	medWR  float64
	minAdv float64
}

func (a AnzHero) String() (str string) {
	str = fmt.Sprintf(`%v:
    Medium Winrate=%.2f%%
    Medium Enemy Winrate=%.2f%%
    Medium Advantage=%+.2f,
    while minimum Advantage=%+.2f`, a.name, a.medWR, a.medWRvs, a.medAdv, a.minAdv)
	return
}

type Anz []AnzHero

func rate(hero AnzHero) float64 {
	return (+hero.medAdv*2 + hero.minAdv*3) / 3
}

func (a Anz) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a Anz) Len() int           { return len(a) }
func (a Anz) Less(i, j int) bool { return rate(a[i]) > rate(a[j]) } //для реверса

func Analize(HeroList map[string]int, AllData [][]vsData, Enemies sort.StringSlice) {
	EC := len(Enemies)
	X := make(Anz, len(HeroList))
	for heroName, heroID := range HeroList {
		X[heroID].name = heroName

		for i, enemy := range Enemies {
			enemyID := HeroList[enemy]
			vsD := AllData[heroID][enemyID]
			X[heroID].medAdv += vsD.Adv
			X[heroID].medWR += vsD.WR
			X[heroID].medWRvs += AllData[enemyID][heroID].WR
			if i == 0 {
				X[heroID].minAdv = vsD.Adv
			} else {
				if vsD.Adv < X[heroID].minAdv {
					X[heroID].minAdv = vsD.Adv
				}
			}
		}
		X[heroID].medAdv /= float64(EC)
		X[heroID].medWR /= float64(EC)
		X[heroID].medWRvs /= float64(EC)
	}
	sort.Sort(X)
	const MaxAnzShow = 5
	showN:=0;
	showloop:
	for _, V := range X {
		for _,e:=range Enemies{
			if e==V.name{
				continue showloop
			}
		}

		if showN > MaxAnzShow-1 {
			break
		}
		fmt.Println("===")
		fmt.Println(showN+1, V)
		showN++
	}
	fmt.Println("===")
}

func showEnemies(Enemies []string) {
	str := ""
	for i, v := range Enemies {
		str += strconv.Itoa(i+1) + ": " + v + " "
	}
	if str == "" {
		str = "no enemy selected"
	}
	fmt.Println("[", str, "]")
}

func terminalprocessor(HeroList map[string]int, AllData [][]vsData) {

	fmt.Println("1) enter a name of hero to add Enemy")
	fmt.Println("2) -dN, ex. d2, deletes selected enemy #2")
	fmt.Println("3) -r - to show result")
	fmt.Println("3) -c - clear all enemies")
	fmt.Println("5) quit or exit - finish")

	var input string
	Enemies := make(sort.StringSlice, 0, 5)
	for {
		Enemies.Sort()
		showEnemies(Enemies)
		fmt.Scanln(&input)
		switch {
		case input == "quit":
			return
		case input == "exit":
			return
		case strings.HasPrefix(input, "-d"):
			num, err := strconv.Atoi(input[2:])
			if err != nil {
				fmt.Println("wrong usage of d command, use dN, ex/ d4")
				continue
			}
			if num < 1 || num > len(Enemies) {
				fmt.Println("no enemy num", num)
				continue
			}
			Enemies = append(Enemies[0:num-1], Enemies[num:]...)
		case input == "-c":
			Enemies = make([]string, 0, 5)
		case input == "-r":
			if len(Enemies) == 0 {
				fmt.Println("no enemies!")
				continue
			}
			Analize(HeroList, AllData, Enemies)
		default: //add new
			if len(Enemies) == 5 {
				fmt.Println("maximum 5 enemies!")
				continue
			}

			variants := make(sort.StringSlice, 0)

		heroloop:
			for heroName, _ := range HeroList {
				if strings.HasPrefix(heroName, input) {
					for _, v := range Enemies {
						if v == heroName {
							continue heroloop
						}
					}
					variants = append(variants, heroName)
				}
			}
			switch len(variants) {
			case 0:
				fmt.Println("no such hero")
			case 1:
				Enemies = append(Enemies, variants[0])
			default: //2+ variants
				variants.Sort()
				for i, v := range variants {
					fmt.Println(i+1, " - ", v)
				}
				fmt.Println("enter number or 0 for back")
				num := 0
				fmt.Scanln(&num)
				if num == 0 {
					continue
				}
				if num < 1 || num > len(variants) {
					fmt.Println("wrong number")
					continue
				}
				Enemies = append(Enemies, variants[num-1])
			}
		}
	}
}
