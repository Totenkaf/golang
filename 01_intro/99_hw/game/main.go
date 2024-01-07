package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"
)

const (
	notImplemented = "Not implemented"
)

func getItemNames(items []*Item) []string {
	names := make([]string, 0, len(items))
	for _, item := range items {
		names = append(names, item.getItemName())
	}
	return names
}

func getBoardInfo(itemsRelations map[*Item][]*Item) []string {
	boardInfo := []string{}
	keys := make([]*Item, 0, len(itemsRelations))
	for key := range itemsRelations {
		keys = append(keys, key)
	}

	sort.SliceStable(keys, func(i, j int) bool {
		return keys[i].getItemName() < keys[j].getItemName()
	})

	for _, key := range keys {
		values := itemsRelations[key]
		if len(values) != 0 && key.getItemName() != "дверь" {
			boardInfo = append(boardInfo, "на "+key.getItemName()+"е: "+strings.Join(getItemNames(values), ", "))
		}
	}
	return boardInfo
}

func getPlaceNames(places []*Place) []string {
	names := make([]string, 0, len(places))
	for _, place := range places {
		names = append(names, place.getPlaceName())
	}
	return names
}

func contains(items []*Item, item *Item) (bool, int) {
	for i, v := range items {
		if v.getItemName() == item.getItemName() {
			return true, i
		}
	}
	return false, -1
}

func remove(items []*Item, idx int) []*Item {
	if idx != -1 {
		return append(items[:idx], items[idx+1:]...)
	}
	return items
}

type Item struct {
	Name      string
	Applyable bool
	Takeable  bool
	Wearable  bool
	Locked    bool
	open      func() string
	apply     func(itemToApply *Item) string
}

func (i Item) getItemName() string {
	return i.Name
}

func (i Item) isTakeable() bool {
	return i.Takeable
}

func (i Item) isWearable() bool {
	return i.Wearable
}

func (i Item) isApplyable() bool {
	return i.Applyable
}

func (i Item) isLocked() bool {
	return i.Locked
}

type Place struct {
	Name           string
	ItemsRelations map[*Item][]*Item
	ActionsToDo    []string
	WaysTo         []*Place
	lookAround     func(pl *Player) string
	move           func(pl *Player) string
}

func (p *Place) setWaysTo(waysTo ...*Place) {
	p.WaysTo = waysTo
}

func (p Place) getWaysTo() []*Place {
	return p.WaysTo
}

func (p Place) checkWayTo(wayTo string) bool {
	for _, way := range p.getWaysTo() {
		if way.getPlaceName() == wayTo {
			return true
		}
	}
	return false
}

func (p Place) getWayByName(wayTo string) *Place {
	for _, way := range p.getWaysTo() {
		if way.getPlaceName() == wayTo {
			return way
		}
	}
	return nil
}

func (p Place) getPlaceName() string {
	return p.Name
}

func (p *Place) addItemsRelations(item *Item, relations ...*Item) {
	p.ItemsRelations[item] = relations
}

func (p Place) getItemsRelations() map[*Item][]*Item {
	return p.ItemsRelations
}

func (p Place) getItems() []*Item {
	items := []*Item{}
	for k, v := range p.ItemsRelations {
		items = append(items, k)
		items = append(items, v...)
	}
	return items
}

func (p Place) isEmpty() bool {
	items := []*Item{}
	for _, v := range p.ItemsRelations {
		items = append(items, v...)
	}
	return len(items) == 0
}

func (p Place) getItemByName(itemName string) *Item {
	for _, item := range p.getItems() {
		if item.getItemName() == itemName {
			return item
		}
	}
	return nil
}

func (p Place) getApplyableItemByName(itemName string) *Item {
	for _, item := range p.getItems() {
		if item.getItemName() == itemName && item.isApplyable() {
			return item
		}
	}
	return nil
}

func (p Place) checkItem(itemName string) bool {
	for _, item := range p.getItems() {
		if item.getItemName() == itemName {
			return true
		}
	}
	return false
}

func (p *Place) removeItem(item *Item) {
	for key, value := range p.ItemsRelations {
		if ok, idx := contains(value, item); ok {
			p.ItemsRelations[key] = remove(value, idx)
		}
	}
}

type Inventory struct {
	Cells  []*Item
	Filled int
}

func (inv *Inventory) addItem(item *Item) {
	if item.isTakeable() {
		inv.Cells = append(inv.Cells, item)
		inv.Filled++
	}
}

func (inv *Inventory) hasItem(itemName string) (bool, int) {
	for i, item := range inv.Cells {
		if item.getItemName() == itemName {
			return true, i
		}
	}
	return false, -1
}

func (inv *Inventory) getItemByName(itemName string) *Item {
	if ok, idx := inv.hasItem(itemName); ok {
		return inv.Cells[idx]
	}
	return nil
}

type Player struct {
	Name      string
	Inv       *Inventory
	CurrPlace *Place
}

func (p Player) getPlayerName() string {
	return p.Name
}

func (p Player) getCurrPlace() *Place {
	return p.CurrPlace
}

func (p *Player) setCurrPlace(place *Place) {
	p.CurrPlace = place
}

func (p *Player) move(newPlace string) string {
	if p.getCurrPlace().checkWayTo(newPlace) {
		newPlaceObj := p.getCurrPlace().getWayByName(newPlace)
		return newPlaceObj.move(p)
	}
	return "нет пути в " + newPlace
}

func (p *Player) useItem(itemName string, itemToApply string) string {
	if ok, _ := p.Inv.hasItem(itemName); ok {
		itemObj := p.Inv.getItemByName(itemName)
		itemToApplyObj := p.getCurrPlace().getApplyableItemByName(itemToApply)
		if itemToApplyObj == nil {
			return "не к чему применить"
		}
		return itemObj.apply(itemToApplyObj)
	}
	return "нет предмета в инвентаре - " + itemName
}

func (p *Player) takeItem(itemName string) string {

	if ok, _ := p.Inv.hasItem("рюкзак"); !ok {
		return "некуда класть"
	}

	if p.CurrPlace.checkItem(itemName) {
		itemObj := p.getCurrPlace().getItemByName(itemName)
		p.Inv.addItem(itemObj)
		p.getCurrPlace().removeItem(itemObj)
		return "предмет добавлен в инвентарь: " + itemObj.getItemName()
	}
	return "нет такого"
}

func (p *Player) wearItem(itemName string) string {
	if p.CurrPlace.checkItem(itemName) {
		itemObj := p.getCurrPlace().getItemByName(itemName)
		if itemObj.isWearable() {
			p.Inv.addItem(itemObj)
			p.getCurrPlace().removeItem(itemObj)
			return "вы надели: " + itemObj.getItemName()
		}
	}
	return "нет такого"
}

func move(gs *GameState, params ...string) string {
	room := params[0]
	return gs.Player.move(room)
}

func lookAround(gs *GameState, params ...string) string {
	flag := params[0]
	if flag == "тихо" {
		return "описание скрыто"
	}
	return gs.Player.getCurrPlace().lookAround(&gs.Player)
}

func useItem(gs *GameState, params ...string) string {
	itemName := params[0]
	itemToApply := params[1]
	return gs.Player.useItem(itemName, itemToApply)
}

func takeItem(gs *GameState, params ...string) string {
	itemName := params[0]
	return gs.Player.takeItem(itemName)
}

func wearItem(gs *GameState, params ...string) string {
	itemName := params[0]
	return gs.Player.wearItem(itemName)
}

type GameState struct {
	Player
	Places []*Place
}

var globalGameState GameState
var globalCommands = map[string]func(gs *GameState, params ...string) string{
	"осмотреться": lookAround,
	"идти":        move,
	"надеть":      wearItem,
	"взять":       takeItem,
	"применить":   useItem,
}

func getPossibleCommands(globalCommands map[string]func(gs *GameState, params ...string) string) string {
	commands := make([]string, 0, len(globalCommands))
	for k := range globalCommands {
		commands = append(commands, k)
	}
	return strings.Join(commands, ", ")
}

func newDoor() *Item {
	door := &Item{
		Name:      "дверь",
		Applyable: true,
		Takeable:  false,
		Wearable:  false,
		Locked:    true,
		apply: func(itemToApply *Item) string {
			return notImplemented
		},
	}
	door.open = func() string {
		if door.Locked {
			door.Locked = false
			return "дверь открыта"
		}
		door.Locked = true
		return "дверь закрыта"
	}
	return door
}

func newTable() *Item {
	return &Item{
		Name:      "стол",
		Applyable: false,
		Takeable:  false,
		Wearable:  false,
		apply: func(itemToApply *Item) string {
			return notImplemented
		},
		open: func() string {
			return notImplemented
		},
	}
}

func initGame() {
	kitchen := &Place{
		Name:           "кухня",
		ItemsRelations: map[*Item][]*Item{},
		WaysTo:         []*Place{},
	}
	kitchen.lookAround = func(pl *Player) string {
		boardInfo := []string{}
		if kitchen.isEmpty() {
			boardInfo = append(boardInfo, "пустая комната")
		} else {
			boardInfo = getBoardInfo(kitchen.getItemsRelations())
		}

		answer := "ты находишься на кухне, "
		answer += strings.Join(boardInfo, ", ") + ", надо "
		if ok, _ := pl.Inv.hasItem("рюкзак"); !ok {
			answer += "собрать рюкзак и "
		}
		answer += "идти в универ. можно пройти - " +
			strings.Join(getPlaceNames(kitchen.getWaysTo()), ", ")
		return answer
	}
	kitchen.move = func(pl *Player) string {
		pl.setCurrPlace(kitchen)
		return kitchen.getPlaceName() +
			", ничего интересного" + ". можно пройти - " +
			strings.Join(getPlaceNames(kitchen.getWaysTo()), ", ")
	}

	hall := &Place{
		Name:           "коридор",
		ItemsRelations: map[*Item][]*Item{},
		WaysTo:         []*Place{},
	}
	hall.lookAround = func(pl *Player) string {
		return "ничего интересного. можно пройти - " +
			strings.Join(getPlaceNames(hall.getWaysTo()), ", ")
	}
	hall.move = func(pl *Player) string {
		pl.setCurrPlace(hall)
		return "ничего интересного. можно пройти - " +
			strings.Join(getPlaceNames(hall.getWaysTo()), ", ")
	}

	myRoom := &Place{
		Name:           "комната",
		ItemsRelations: map[*Item][]*Item{},
		WaysTo:         []*Place{},
	}
	myRoom.lookAround = func(pl *Player) string {
		boardInfo := []string{}
		if myRoom.isEmpty() {
			boardInfo = append(boardInfo, "пустая комната")
		} else {
			boardInfo = getBoardInfo(myRoom.getItemsRelations())
		}
		return strings.Join(boardInfo, ", ") + ". можно пройти - " +
			strings.Join(getPlaceNames(myRoom.getWaysTo()), ", ")
	}
	myRoom.move = func(pl *Player) string {
		pl.setCurrPlace(myRoom)
		return "ты в своей комнате. можно пройти - " +
			strings.Join(getPlaceNames(myRoom.getWaysTo()), ", ")
	}

	street := &Place{
		Name:           "улица",
		ItemsRelations: map[*Item][]*Item{},
		WaysTo:         []*Place{},
		lookAround: func(pl *Player) string {
			return notImplemented
		},
	}
	street.move = func(pl *Player) string {
		doorObj := pl.getCurrPlace().getItemByName("дверь")
		if doorObj.isLocked() {
			return "дверь закрыта"
		}
		pl.setCurrPlace(street)
		return "на улице весна. можно пройти - " + street.getWaysTo()[0].getPlaceName() + "ой"
	}

	home := &Place{
		Name:           "дом",
		ItemsRelations: map[*Item][]*Item{},
		WaysTo:         []*Place{},
		lookAround: func(pl *Player) string {
			return notImplemented
		},
		move: func(pl *Player) string {
			return notImplemented
		},
	}

	kitchen.setWaysTo(hall)
	hall.setWaysTo(kitchen, myRoom, street)
	myRoom.setWaysTo(hall)
	street.setWaysTo(home)

	tea := &Item{
		Name:      "чай",
		Applyable: false,
		Takeable:  false,
		Wearable:  false,
		apply: func(itemToApply *Item) string {
			return notImplemented
		},
		open: func() string {
			return notImplemented
		},
	}
	key := &Item{
		Name:      "ключи",
		Applyable: true,
		Takeable:  true,
		apply: func(itemToApply *Item) string {
			if itemToApply.getItemName() == "дверь" {
				return itemToApply.open()
			}
			return "не к чему применить"
		},
		open: func() string {
			return notImplemented
		},
	}
	conspects := &Item{
		Name:      "конспекты",
		Applyable: false,
		Takeable:  true,
		Wearable:  false,
		apply: func(itemToApply *Item) string {
			return notImplemented
		},
		open: func() string {
			return notImplemented
		},
	}
	bag := &Item{
		Name:      "рюкзак",
		Applyable: false,
		Takeable:  true,
		Wearable:  true,
		apply: func(itemToApply *Item) string {
			return notImplemented
		},
		open: func() string {
			return notImplemented
		},
	}
	chair := &Item{
		Name:      "стул",
		Applyable: false,
		Takeable:  false,
		Wearable:  false,
		apply: func(itemToApply *Item) string {
			return notImplemented
		},
		open: func() string {
			return notImplemented
		},
	}

	kitchen.addItemsRelations(newDoor())
	kitchen.addItemsRelations(newTable(), tea)

	myRoom.addItemsRelations(newDoor())
	myRoom.addItemsRelations(newTable(), key, conspects)
	myRoom.addItemsRelations(chair, bag)

	hall.addItemsRelations(newDoor())

	globalGameState = GameState{
		Player: Player{
			Name: "Player",
			Inv: &Inventory{
				Cells:  []*Item{},
				Filled: 0,
			},
		},
		Places: []*Place{
			kitchen,
			hall,
			myRoom,
			street,
			home,
		},
	}
	globalGameState.Player.setCurrPlace(kitchen)
}

func (g *GameState) changeStateByCommand(input string) string {
	argv := strings.Split(input+" ", " ")
	command := argv[0]
	params := argv[1:]

	if function, ok := globalCommands[command]; ok {
		return function(g, params...)
	} else {
		return "неизвестная команда"
	}
}

func handleCommand(command string) string {
	return globalGameState.changeStateByCommand(command)
}

func main() {
	reader := bufio.NewReader(os.Stdin)

	initGame()

	fmt.Println("\n==========")
	fmt.Printf("Player name: %s\n", globalGameState.Player.getPlayerName())
	fmt.Printf("Player start postion: %s\n", globalGameState.Player.getCurrPlace().getPlaceName())
	fmt.Println("==========")

	commands := getPossibleCommands(globalCommands)
	for {
		fmt.Println("\n==========")
		fmt.Printf("Possible commands: %s\n", commands)
		fmt.Println("==========")

		fmt.Print("\nEnter comand: \n")
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Print("Unexpected error. Programm stopped")
			return
		}
		input = strings.ReplaceAll(input, "\n", "")
		message := handleCommand(input)
		fmt.Println(message)
	}
}
