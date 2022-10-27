package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/foolin/mixer"
	"github.com/maxence-charriere/go-app/v9/pkg/app"
	shell "github.com/stateless-minds/go-ipfs-api"
)

const dbAddressSupplyDemand = "/orbitdb/bafyreia6t57n2uyfgpwpjqsztoxiiluc5xwcidfrfulc4i2quyd65uhmpe/demand_supply"

const (
	topicDemand   = "demand"
	topicCritical = "critical"
)
const (
	Hour                                   = "hour"
	Day                                    = "day"
	Week                                   = "week"
	Month                                  = "month"
	Year                                   = "year"
	Custom                                 = "custom"
	NotificationSuccess NotificationStatus = "success"
	NotificationInfo    NotificationStatus = "info"
	NotificationWarning NotificationStatus = "warning"
	NotificationDanger  NotificationStatus = "danger"
	NotificationPrimary NotificationStatus = "primary"
)

// pubsub is a component that does a simple pubsub on ipfs. A component is a
// customizable, independent, and reusable UI element. It is created by
// embedding app.Compo into a struct.
type pubsub struct {
	app.Compo
	topic string
	demandRequest
	sendRequest
	demandRequests           map[string]demandRequest
	categories               []string
	sh                       *shell.Shell
	sub                      *shell.PubSubSubscription
	citizenID                string
	newComer                 bool
	timeRange                []int
	timeFormat               []string
	ratio                    []int
	index                    []int
	hasChildren              bool
	filteredRequests         []int
	validRequests            []int
	filteredandValidRequests int
	filteredWaterRequests    []int
	filteredFoodRequests     []int
	filteredHousingRequests  []int
	filteredOtherRequests    []int
	lastWaterRequest         int
	lastFoodRequest          int
	lastHousingRequest       int
	lastOtherRequest         int
	ranks                    []ranking
	showMessages             bool
	showRatio                bool
	showTime                 bool
	showChart                bool
	showRanks                bool
	counterDemand            int
	counterSupply            int
	counterSameTime          int
	counterSameTimeFulfilled int
	coordinates              []coordinate
	category                 string
	period                   string
	stats                    string
	multiplyer               int
	notifications            map[string]notification
	notificationID           int
	globalEvent              bool
	resource                 string
}

type NotificationStatus string

type notification struct {
	id      int
	status  string
	header  string
	message string
}

type coordinate struct {
	id       int
	top      int
	left     int
	angle    float64
	distance float64
}

type ranking struct {
	citizenID       string
	demandRatio     float64 // number of personal demands compared to total
	supplyRatio     float64 // number of personal supplies compared to total
	reputationIndex float64 // supplyRatio compared to demandRatio
}

type sendRequest struct {
	ID          int
	Category    string
	Quantity    string
	Details     string
	Fulfilled   bool
	CreatedAt   time.Time
	FulfilledAt time.Time
}

type demandRequest struct {
	ID          int
	CitizenID   string
	Category    string
	Quantity    string
	Details     string
	CreatedAt   time.Time
	Fulfilled   bool
	FulfilledBy string
	FulfilledAt time.Time
}

func (p *pubsub) OnMount(ctx app.Context) {
	p.topic = topicDemand
	sh := shell.NewShell("localhost:5001")
	p.sh = sh
	myPeer, err := p.sh.ID()
	if err != nil {
		log.Fatal(err)
	}

	citizenID := myPeer.ID[len(myPeer.ID)-8:]
	// replace password with your own
	password := "mysecretpassword"

	p.citizenID = mixer.EncodeString(password, citizenID)

	p.subscribe(ctx)
	p.demandRequests = make(map[string]demandRequest)
	p.FetchAllRequests(ctx, app.Event{})
	p.setTimeAxis(Hour)
	// 0 to 1 supply/demand
	for i := 1; i < 11; i++ {
		p.ratio = append(p.ratio, i)
	}
	// default categories
	p.categories = append(p.categories, "all", "water", "food", "housing", "other")
	p.category = "All"
	p.showRatio = true
	p.showTime = true
	p.showRanks = false
	// create welcome notification
	p.notifications = make(map[string]notification)
}

func (p *pubsub) setTimeAxis(period string) {
	p.period = period
	p.timeRange = []int{}
	p.timeFormat = []string{}
	switch period {
	case "hour":
		now := time.Now()
		// defining duration
		d := (10 * time.Minute)
		now = now.Round(d)
		before := now.Add(time.Hour * -1)
		for i := 1; i < 7; i++ {
			p.timeRange = append(p.timeRange, i)
			before = before.Add(time.Minute * 10)
			min := strconv.Itoa(before.Minute())
			if before.Minute() == 0 {
				min += "0"
			}
			p.timeFormat = append(p.timeFormat, strconv.Itoa(before.Hour())+":"+min)
		}
		p.multiplyer = 132
	case "day":
		for i := 1; i < 25; i++ {
			p.timeRange = append(p.timeRange, i)
			p.timeFormat = append(p.timeFormat, strconv.Itoa(i)+":00")
		}
		p.multiplyer = 33
	case "week":
		now := time.Now()
		n := 8
		for i := 1; i < n; i++ {
			p.timeRange = append(p.timeRange, i)
			before := now.AddDate(0, 0, i+1-n)
			p.timeFormat = append(p.timeFormat, strconv.Itoa(before.Day())+". "+time.Now().Month().String())
		}
		p.multiplyer = 113
	case "month":
		before := time.Now().AddDate(0, 0, -30)

		for i := 1; i < 31; i++ {
			p.timeRange = append(p.timeRange, i)
			before = before.AddDate(0, 0, 1)
			p.timeFormat = append(p.timeFormat, strconv.Itoa(before.Day())+". "+before.Month().String())
		}
		p.multiplyer = 26
	case "year":
		now := time.Now()
		before := now.AddDate(-1, 0, 0)
		for i := 1; i < 13; i++ {
			p.timeRange = append(p.timeRange, i)
			before = before.AddDate(0, 1, 0)
			p.timeFormat = append(p.timeFormat, before.Month().String()+". "+strconv.Itoa(before.Year()))
		}
		p.multiplyer = 66
	}
}

// The Render method is where the component appearance is defined. Here, a
// "pubsub World!" is displayed as a heading.
func (p *pubsub) Render() app.UI {
	return app.Div().Class("container").Body(
		app.Link().Rel("stylesheet").Href("https://cdn.jsdelivr.net/npm/bootstrap@5.0.2/dist/css/bootstrap.min.css").CrossOrigin("anonymous"),
		app.Script().Src("https://cdn.jsdelivr.net/npm/bootstrap@5.0.2/dist/js/bootstrap.bundle.min.js").CrossOrigin("anonymous"),
		app.Div().Class("square_box box_three"),
		app.Div().Class("square_box box_four"),
		app.If(len(p.notifications) > 0,
			app.Section().Body(
				app.Div().Class("container").Body(
					app.Div().Class("row").Body(
						app.Range(p.notifications).Map(func(s string) app.UI {
							return app.Div().Class("col-sm-12").Body(
								app.Div().Class("alert fade alert-simple alert-"+p.notifications[s].status+" alert-dismissible text-left font__family-montserrat font__size-16 font__weight-light brk-library-rendered rendered show").Body(
									app.Button().ID("notify-"+strconv.Itoa(p.notifications[s].id)).Class("btn-close btn-close-white").Type("button").DataSet("bs-dismiss", "alert").Aria("label", "Close").Value(p.notifications[s].id),
									app.I().Class("start-icon far fa-check-circle faa-tada animated"),
									app.Strong().Class("font__weight-semibold").Text(p.notifications[s].header+" "),
									app.Text(p.notifications[s].message),
								),
							)
						}),
					),
				),
			)),
		app.H1().Class("pb-0 logo").Body(
			app.Text("Cyber-Stasis"),
			app.Details().Body(
				app.Summary().Body(
					app.Small().Body(
						app.Div().Class("button").Text("How to Play"),
					),
					app.Div().Class("details-modal-overlay"),
				),
				app.Div().Class("details-modal").Body(
					app.Div().Class("details-modal-close").Body(
						app.Button().Class("btn-close btn-sm").Type("button").Aria("label", "Close"),
					),
					app.Div().Class("details-modal-title").Body(
						app.H1().Text("How to Play"),
					),
					app.Div().Class("details-modal-content").Body(
						app.Div().Body(
							app.Details().Class("how-to-play").Open(true).Body(
								app.Summary().Class("accordion").Text("The Story"),
								app.Div().Text("Welcome people from the past. I am here to tell you what happened after the Great Reset. The year is 2105. Humanity now lives in a post-capitalism era. We no longer work by schedule and necessity. Automation has replaced all repetitive and dangerous jobs. The concept of work has diminished. There are no countries and the concept of money has long been forgotten. Private property was abolished and corporations were transformed into cooperatives. We are all connected in an open source p2p network. Every day we open up our Cyber Stasis dashboards to request goods and services we need for the day and to provide what we can."),
							),
							app.Details().Class("how-to-play").Body(
								app.Summary().Class("accordion").Text("The Mission"),
								app.Div().Text("The mission of Cyber Stasis is to prototype a working economic simulator as an alternative to the monetary system. It's goal is to be a living proof that the market economy can operate without exchange. Recent advancement in p2p technology made possible the release of the Cyber Stasis dashboard as an open-source technology which is free and hosted by everyone who uses it. This allows the project to run forever as long as there are people participating. Hopefully this experiment can help academic institutions to research alternative economic models further for the betterment of society as a whole."),
							),
							app.Details().Class("how-to-play").Body(
								app.Summary().Class("accordion").Text("Guidelines"),
								app.Div().Body(
									app.Ol().Class("list-group list-group-numbered").Body(
										app.Li().Class("list-group-item d-flex justify-content-between align-items-start").Body(
											app.Div().Class("ms-2 me-auto").Body(
												app.Div().Class("fw-bold").Text("Economic simulator"),
												app.Text("Cyber Stasis is an economic simulator in the form of a fictional game based on global real-time demand and supply."),
											),
										),
										app.Li().Class("list-group-item d-flex justify-content-between align-items-start").Body(
											app.Div().Class("ms-2 me-auto").Body(
												app.Div().Class("fw-bold").Text("Real-time demand/supply graph"),
												app.Text("The graph reflects all demand and supply requests and is updated in real-time."),
											),
										),
										app.Li().Class("list-group-item d-flex justify-content-between align-items-start").Body(
											app.Div().Class("ms-2 me-auto").Body(
												app.Div().Class("fw-bold").Text("Supply can be sent only in response to an existing demand"),
												app.Text("Send only goods and services you can provide in real life."),
											),
										),
										app.Li().Class("list-group-item d-flex justify-content-between align-items-start").Body(
											app.Div().Class("ms-2 me-auto").Body(
												app.Div().Class("fw-bold").Text("Keep it real"),
												app.Text("Send requests for your real daily needs to make the whole simulation as accurate as possible."),
											),
										),
										app.Li().Class("list-group-item d-flex justify-content-between align-items-start").Body(
											app.Div().Class("ms-2 me-auto").Body(
												app.Div().Class("fw-bold").Text("Global events"),
												app.Text("When the supply/demand ratio drops below certain thresholds global events are triggered and sent as notifications such as global shortage of water, food and housing."),
											),
										),
										app.Li().Class("list-group-item d-flex justify-content-between align-items-start").Body(
											app.Div().Class("ms-2 me-auto").Body(
												app.Div().Class("fw-bold").Text("Do what you do in real life"),
												app.Text("Ask for things you need and supply things you provide."),
											),
										),
										app.Li().Class("list-group-item d-flex justify-content-between align-items-start").Body(
											app.Div().Class("ms-2 me-auto").Body(
												app.Div().Class("fw-bold").Text("Rankings"),
												app.Text("Rankings reflect the level of contribution and usefulness of members to society. They take all factors into account and are calculated by a formula. The Reputation Index is the score in the game. Provide more than you consume and become the most valuable member of society!"),
											),
										),
										app.Li().Class("list-group-item d-flex justify-content-between align-items-start").Body(
											app.Div().Class("ms-2 me-auto").Body(
												app.Div().Class("fw-bold").Text("No collection of user data"),
												app.Text("Cyber Stasis does not collect any personal user data."),
											),
										),
										app.Li().Class("list-group-item d-flex justify-content-between align-items-start").Body(
											app.Div().Class("ms-2 me-auto").Body(
												app.Div().Class("fw-bold").Text("User generated content is fictional"),
												app.Text("All user generated content is fictional and creators are not responsibile for it."),
											),
										),
										app.Li().Class("list-group-item d-flex justify-content-between align-items-start").Body(
											app.Div().Class("ms-2 me-auto").Body(
												app.Div().Class("fw-bold").Text("If you like it support it"),
												app.Text("This is an open source community project. Feel free to improve it or fork it and use it for your projects. Donations are welcome."),
											),
										),
									),
								),
							),
							app.Details().Class("how-to-play").Body(
								app.Summary().Class("accordion").Text("Inspirations"),
								app.Div().Body(
									app.Ol().Class("list-group list-group-numbered").Body(
										app.Li().Class("list-group-item d-flex justify-content-between align-items-start").Body(
											app.Div().Class("ms-2 me-auto").Body(
												app.Div().Class("fw-bold").Text("Auroville"),
												app.A().Href("https://auroville-learning.net").Text("https://auroville-learning.net"),
											),
										),
										app.Li().Class("list-group-item d-flex justify-content-between align-items-start").Body(
											app.Div().Class("ms-2 me-auto").Body(
												app.Div().Class("fw-bold").Text("CyberSyn"),
												app.A().Href("https://en.wikipedia.org/wiki/Project_Cybersyn").Text("https://en.wikipedia.org/wiki/Project_Cybersyn"),
											),
										),
										app.Li().Class("list-group-item d-flex justify-content-between align-items-start").Body(
											app.Div().Class("ms-2 me-auto").Body(
												app.Div().Class("fw-bold").Text("The Venus Project"),
												app.A().Href("https://www.thevenusproject.com").Text("https://www.thevenusproject.com"),
											),
										),
										app.Li().Class("list-group-item d-flex justify-content-between align-items-start").Body(
											app.Div().Class("ms-2 me-auto").Body(
												app.Div().Class("fw-bold").Text("OGAS"),
												app.A().Href("https://en.wikipedia.org/wiki/OGAS").Text("https://en.wikipedia.org/wiki/OGAS"),
											),
										),
									),
								),
							),
							app.Details().Class("how-to-play").Body(
								app.Summary().Class("accordion").Text("Acknowledgments"),
								app.Div().Body(
									app.Ol().Class("list-group list-group-numbered").Body(
										app.Li().Class("list-group-item d-flex justify-content-between align-items-start").Body(
											app.Div().Class("ms-2 me-auto").Body(
												app.Div().Class("fw-bold").Text("IPFS"),
												app.A().Href("https://ipfs.io").Text("https://ipfs.io"),
											),
										),
										app.Li().Class("list-group-item d-flex justify-content-between align-items-start").Body(
											app.Div().Class("ms-2 me-auto").Body(
												app.Div().Class("fw-bold").Text("Berty"),
												app.A().Href("https://berty.tech").Text("https://berty.tech"),
											),
										),
										app.Li().Class("list-group-item d-flex justify-content-between align-items-start").Body(
											app.Div().Class("ms-2 me-auto").Body(
												app.Div().Class("fw-bold").Text("GO-APP"),
												app.A().Href("https://go-app.dev").Text("https://go-app.dev"),
											),
										),
										app.Li().Class("list-group-item d-flex justify-content-between align-items-start").Body(
											app.Div().Class("ms-2 me-auto").Body(
												app.Div().Class("fw-bold").Text("Alessio Atzeni"),
												app.A().Href("https://www.alessioatzeni.com/blog/css3-graph-animation").Text("https://www.alessioatzeni.com/blog/css3-graph-animation"),
											),
										),
										app.Li().Class("list-group-item d-flex justify-content-between align-items-start").Body(
											app.Div().Class("ms-2 me-auto").Body(
												app.Div().Class("fw-bold").Text("Niels Voogt"),
												app.A().Href("https://codepen.io/NielsVoogt/pen/XWjPdjO").Text("https://codepen.io/NielsVoogt/pen/XWjPdjO"),
											),
										),
										app.Li().Class("list-group-item d-flex justify-content-between align-items-start").Body(
											app.Div().Class("ms-2 me-auto").Body(
												app.Div().Class("fw-bold").Text("Håvard Brynjulfsen"),
												app.A().Href("https://codepen.io/havardob/pen/abBJgQo").Text("https://codepen.io/havardob/pen/abBJgQo"),
											),
										),
									),
								),
							),
							app.Details().Class("how-to-play").Body(
								app.Summary().Class("accordion").Text("Support Us"),
								app.Div().Body(
									app.A().Href("https://opencollective.com/stateless-minds-collective").Text("https://opencollective.com/stateless-minds-collective"),
								),
							),
							app.Details().Class("how-to-play").Body(
								app.Summary().Class("accordion").Text("Terms of Service"),
								app.Div().Body(
									app.Ul().Class("list-group").Body(
										app.Li().Class("list-group-item d-flex justify-content-between align-items-start").Body(
											app.Div().Class("ms-2 me-auto").Body(
												app.Div().Class("fw-bold").Text("Introduction"),
												app.Text("Cyber Stasis is an economic simulator in the form of a fictional game based on global real-time demand and supply. By using the application you are implicitly agreeing to share your peer id with the IPFS public network."),
											),
										),
										app.Li().Class("list-group-item d-flex justify-content-between align-items-start").Body(
											app.Div().Class("ms-2 me-auto").Body(
												app.Div().Class("fw-bold").Text("Application Hosting"),
												app.Text("Cyber Stasis is a decentralized application and is hosted on a public peer to peer network. By using the application you agree to host it on the public IPFS network free of charge for as long as your usage is."),
											),
										),
										app.Li().Class("list-group-item d-flex justify-content-between align-items-start").Body(
											app.Div().Class("ms-2 me-auto").Body(
												app.Div().Class("fw-bold").Text("User-Generated Content"),
												app.Text("All published content is user-generated, fictional and creators are not responsible for it."),
											),
										),
									),
								),
							),
							app.Details().Class("how-to-play").Body(
								app.Summary().Class("accordion").Text("Privacy Policy"),
								app.Div().Body(
									app.Ul().Class("list-group").Body(
										app.Li().Class("list-group-item d-flex justify-content-between align-items-start").Body(
											app.Div().Class("ms-2 me-auto").Body(
												app.Div().Class("fw-bold").Text("Personal Data"),
												app.Text("There is no personal information collected within Cyber Stasis. We store a small portion of your peer ID encrypted as a non-unique identifier which is used for displaying the ranks interface."),
											),
										),
										app.Li().Class("list-group-item d-flex justify-content-between align-items-start").Body(
											app.Div().Class("ms-2 me-auto").Body(
												app.Div().Class("fw-bold").Text("Cookies"),
												app.Text("Cyber Stasis does not use cookies."),
											),
										),
										app.Li().Class("list-group-item d-flex justify-content-between align-items-start").Body(
											app.Div().Class("ms-2 me-auto").Body(
												app.Div().Class("fw-bold").Text("Links to External Websites"),
												app.Text("Cyber Stasis contains links to other websites. If you click on a third-party link, you will be directed to that site. Note that these external sites are not operated by Cyber Stasis. Therefore, it is strongly advised that you review the Privacy Policy of these websites. Cyber Stasis has no control over and assumes no responsibility for the content, privacy policies, or practices of any third-party sites or services."),
											),
										),
										app.Li().Class("list-group-item d-flex justify-content-between align-items-start").Body(
											app.Div().Class("ms-2 me-auto").Body(
												app.Div().Class("fw-bold").Text("Changes to this Privacy Policy"),
												app.Text("This Privacy Policy might be updated from time to time. Thus, it is advised to review this page periodically for any changes. You will be notified of any changes from this page. Changes are effective immediately after they are posted on this page."),
											),
										),
									),
								),
							),
						),
					),
				),
			),
		),
		app.Div().Class("container d-flex justify-content-center pb-5").Body(
			app.Div().Class("card").Body(
				app.If(p.showMessages,
					app.H6().Class("card-title").Text("Pending Requests"),
					app.Div().Class("card-body").Body(
						app.Range(p.index).Slice(func(i int) app.UI {
							i = p.index[i]
							if p.demandRequests[strconv.Itoa(i)].ID > 0 && !p.demandRequests[strconv.Itoa(i)].Fulfilled {
								return app.Div().Class("d-flex flex-row p-3").Body(
									app.Img().Src("https://img.icons8.com/color/48/000000/circled-user-female-skin-type-7.png").Width(30).Height(30),
									app.Div().Class("chat ml-3 p-3").Body(
										app.Span().Class("pe-2").Body(
											app.Span().Class("badge rounded-pill bg-info text-dark").Text(strings.ToUpper(p.demandRequests[strconv.Itoa(i)].Category)),
											app.P().Class("card-text pt-3").Text("Quantity: "+p.demandRequests[strconv.Itoa(i)].Quantity),
											app.P().Class("card-text").Text("Details: "+p.demandRequests[strconv.Itoa(i)].Details),
											app.Div().Class("row d-flex justify-content-center align-content-center ps-3 pe-3").Body(
												app.Button().Class("btn btn-outline-primary btn-sm rounded-pill").ID(strconv.Itoa(p.demandRequests[strconv.Itoa(i)].ID)).Body(app.Text("Send Supply")).OnClick(p.sendSupply),
											),
										),
									),
								)
							}
							return nil
						})),
				),
				app.H6().Class("card-title").Text("What do you need today?"),
				app.Div().Class("form-group").Body(
					app.Select().Class("form-select").Aria("label", "Demand category").Body(
						app.Range(p.categories).Slice(func(i int) app.UI {
							if p.categories[i] == "all" {
								return app.Option().Selected(true).Value("").Text("Select Category")
							} else {
								return app.Option().Value(p.categories[i]).Text(strings.Title(p.categories[i]))
							}
						}),
					).Required(true).OnClick(p.onSelect),
					app.Input().ID("quantity").Class("form-control").Name("quantity").Type("number").Placeholder("Quantity").OnKeyUp(p.onInput),
					app.Textarea().Class("form-control").Rows(3).Placeholder("Details").OnKeyUp(p.onMessage),
				),
				app.Button().Class("btn btn-outline-info mt-2").ID("submitDemand").Body(app.Text("Send Request")).OnClick(p.sendDemand).Disabled(true),
				// app.Button().Class("btn btn-outline-secondary").ID("FetchAllRequests").Body(app.Text("Get Requests")).OnClick(p.FetchAllRequests),
				app.Button().Class("btn btn-outline-warning").ID("dummydata").Body(app.Text("Dummy Data")).OnClick(p.dummyData),
				// app.Button().Class("btn btn-outline-danger").ID("deleteRequests").Body(app.Text("Delete Requests")).OnClick(p.deleteRequests),
			),
			app.Div().ID("secondary").Class("container").Body(
				app.Button().Class("btn btn-outline-info category").Text("Other").Value("Other").OnClick(p.onSelectCategory),
				app.Button().Class("btn btn-outline-info category").Text("Housing").Value("Housing").OnClick(p.onSelectCategory),
				app.Button().Class("btn btn-outline-info category").Text("Food").Value("Food").OnClick(p.onSelectCategory),
				app.Button().Class("btn btn-outline-info category").Text("Water").Value("Water").OnClick(p.onSelectCategory),
				app.Button().ID("category-all").Class("btn btn-outline-info category active").Text("All").Value("All").OnClick(p.onSelectCategory),
				app.Button().ID("global-stats").Class("btn btn-outline-info stats active").Text("Global Stats").Value("Global").OnClick(p.onSelectStats),
				app.Button().ID("ranks").Class("btn btn-outline-info ranks").Text("Ranks").Value("Ranks").OnClick(p.onSelectRanks),
				app.Button().Class("btn btn-outline-info period").Text("1 Year").Value(Year).OnClick(p.onSelectPeriod),
				app.Button().Class("btn btn-outline-info period").Text("1 Month").Value(Month).OnClick(p.onSelectPeriod),
				app.Button().Class("btn btn-outline-info period").Text("1 Week").Value(Week).OnClick(p.onSelectPeriod),
				app.Button().Class("btn btn-outline-info period").Text("1 Day").Value(Day).OnClick(p.onSelectPeriod),
				app.Button().ID("period-hour").Class("btn btn-outline-info period active").Text("1 Hour").Value(Hour).OnClick(p.onSelectPeriod),
				app.Button().ID("my-stats").Class("btn btn-outline-info stats").Text("My Stats").Value("Personal").OnClick(p.onSelectStats),
				app.Div().Class("content running").Body(
					app.If(p.showRanks, app.Ol().Class("list-group list-group-numbered").Body(
						app.Range(p.ranks).Slice(func(i int) app.UI {
							var class string
							if p.ranks[i].citizenID == p.citizenID {
								class = "active"
							}
							if p.ranks[i].citizenID == "" {
								return nil
							}

							return app.Li().Class("list-group-item "+class+" d-flex justify-content-between align-items-start").Body(
								app.Div().Class("ms-2 me-auto").Body(
									app.Div().Class("fw-bold").Text("Citizen"),
									app.Span().Class("badge bg-primary rounded-pill").Text(p.ranks[i].citizenID),
								),
								app.Div().Class("ms-2 me-auto").Body(
									app.Div().Class("fw-bold").Text("Demand"),
									app.Span().Class("badge bg-primary rounded-pill").Text(p.ranks[i].demandRatio),
								),
								app.Div().Class("ms-2 me-auto").Body(
									app.Div().Class("fw-bold").Text("Supply"),
									app.Span().Class("badge bg-primary rounded-pill").Text(p.ranks[i].supplyRatio),
								),
								app.Div().Class("ms-2 me-auto").Body(
									app.Div().Class("fw-bold").Text("Reputation"),
									app.Span().Class("badge bg-primary rounded-pill").Text(p.ranks[i].reputationIndex),
								),
							)
						}),
					)),
					app.If(p.showRatio,
						app.Range(p.ratio).Slice(func(i int) app.UI {
							return app.Div().Class("range").Style("top", strconv.Itoa(390-(p.ratio[i]*40))+"px").Style("left", "0").Body(
								app.A().Href("#").Body(app.Small().Text("Supply/Demand Ratio: " + fmt.Sprintf("%.1f", float32(p.ratio[i])/10))),
							)
						}),
					),
					app.If(p.showTime,
						app.Range(p.timeRange).Slice(func(i int) app.UI {
							return app.Div().Class("time").Style("top", "390px").Style("left", strconv.Itoa(p.timeRange[i]*p.multiplyer)+"px").Body(
								app.A().Href("#").Body(app.Small().Text(p.timeFormat[i])),
							)
						}),
					),
					app.If(p.showChart,
						app.If(len(p.filteredRequests) > 0,
							app.Range(p.filteredRequests).Slice(func(i int) app.UI {
								if p.stats == "Personal" && p.demandRequests[strconv.Itoa(p.filteredRequests[i])].CitizenID != p.citizenID {
									return nil
								}

								t := time.Now()
								// defining duration
								d := (10 * time.Minute)
								t = t.Round(d)

								var left int

								switch p.period {
								case "hour":
									index := 6
									tt := t.Sub(p.demandRequests[strconv.Itoa(p.filteredRequests[i])].CreatedAt)
									if tt.Minutes() >= 60 || tt.Minutes() < 0 {
										p.filteredandValidRequests--
										return nil
									}

									if p.filteredandValidRequests < 2 {
										return nil
									}

									p.validRequests = append(p.validRequests, p.filteredRequests[i])
									if p.filteredRequests[i] == p.validRequests[0] {
										for n := 1; n < 6; n++ {
											switch m := tt.Minutes(); {
											case int(m) >= 10*n:
												index--
											}
										}

										p.counterSameTime++
										if p.demandRequests[strconv.Itoa(p.filteredRequests[i])].Fulfilled {
											p.counterSameTimeFulfilled++
										}

										if p.demandRequests[strconv.Itoa(p.filteredRequests[i+1])].CreatedAt.Sub(p.demandRequests[strconv.Itoa(p.filteredRequests[i])].CreatedAt).Minutes() < 10 {
											return nil
										}
									} else {
										if p.demandRequests[strconv.Itoa(p.filteredRequests[i])].CreatedAt.Sub(p.demandRequests[strconv.Itoa(p.filteredRequests[i-1])].CreatedAt).Minutes() < 10 {
											p.counterSameTime++
											if p.demandRequests[strconv.Itoa(p.filteredRequests[i])].Fulfilled {
												p.counterSameTimeFulfilled++
											}
											p.hasChildren = true
											return nil
										}

										var ttt time.Duration

										if p.hasChildren {
											ttt = t.Sub(p.demandRequests[strconv.Itoa(p.filteredRequests[i-1])].CreatedAt)
										} else {
											ttt = tt
										}

										for n := 1; n < 6; n++ {
											switch m := ttt.Minutes(); {
											case int(m) >= 10*n:
												index--
											}
										}
									}

									left = index * p.multiplyer
								case "day":
									index := 24
									t = time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 24, 0, 0, 0, time.Local)
									tt := t.Sub(p.demandRequests[strconv.Itoa(p.filteredRequests[i])].CreatedAt)
									if tt.Hours() > 24 || tt.Hours() < 0 {
										p.filteredandValidRequests--
										return nil
									}

									if p.filteredandValidRequests < 2 {
										return nil
									}

									p.validRequests = append(p.validRequests, p.filteredRequests[i])

									if p.filteredRequests[i] == p.validRequests[0] {
										index = p.demandRequests[strconv.Itoa(p.filteredRequests[i])].CreatedAt.Hour()
										p.counterSameTime++
										if p.demandRequests[strconv.Itoa(p.filteredRequests[i])].Fulfilled {
											p.counterSameTimeFulfilled++
										}

										if p.demandRequests[strconv.Itoa(p.filteredRequests[i+1])].CreatedAt.Sub(p.demandRequests[strconv.Itoa(p.filteredRequests[i])].CreatedAt).Hours() < 1 {
											return nil
										}
									} else {
										index = p.demandRequests[strconv.Itoa(p.filteredRequests[i-1])].CreatedAt.Hour()
										if p.demandRequests[strconv.Itoa(p.filteredRequests[i])].CreatedAt.Sub(p.demandRequests[strconv.Itoa(p.filteredRequests[i-1])].CreatedAt).Hours() < 1 {
											p.counterSameTime++
											if p.demandRequests[strconv.Itoa(p.filteredRequests[i])].Fulfilled {
												p.counterSameTimeFulfilled++
											}
											p.hasChildren = true
											return nil
										} else {
											if !p.hasChildren {
												index = p.demandRequests[strconv.Itoa(p.filteredRequests[i])].CreatedAt.Hour()
											}
										}
									}

									left = index * p.multiplyer
								case "week":
									index := 7
									t = time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 24, 0, 0, 0, time.Local)
									tt := t.Sub(p.demandRequests[strconv.Itoa(p.filteredRequests[i])].CreatedAt)
									if tt.Hours() > 168 || tt.Hours() < 0 {
										p.filteredandValidRequests--
										return nil
									}

									if p.filteredandValidRequests < 2 {
										return nil
									}

									p.validRequests = append(p.validRequests, p.filteredRequests[i])

									if p.filteredRequests[i] == p.validRequests[0] {
										for n := 1; n < 7; n++ {
											switch h := tt.Hours(); {
											case int(h) > 24*n:
												index--
											}
										}
										p.counterSameTime++
										if p.demandRequests[strconv.Itoa(p.filteredRequests[i])].Fulfilled {
											p.counterSameTimeFulfilled++
										}

										if p.demandRequests[strconv.Itoa(p.filteredRequests[i+1])].CreatedAt.Sub(p.demandRequests[strconv.Itoa(p.filteredRequests[i])].CreatedAt).Hours() < 24 {
											return nil
										}
									} else {
										if p.demandRequests[strconv.Itoa(p.filteredRequests[i])].CreatedAt.Sub(p.demandRequests[strconv.Itoa(p.filteredRequests[i-1])].CreatedAt).Hours() < 24 {
											p.counterSameTime++
											if p.demandRequests[strconv.Itoa(p.filteredRequests[i])].Fulfilled {
												p.counterSameTimeFulfilled++
											}
											p.hasChildren = true
											return nil
										}

										var ttt time.Duration

										if p.hasChildren {
											ttt = t.Sub(p.demandRequests[strconv.Itoa(p.filteredRequests[i-1])].CreatedAt)
										} else {
											ttt = tt
											if p.demandRequests[strconv.Itoa(p.filteredRequests[i])].Fulfilled {
												p.counterSameTimeFulfilled++
											}
										}

										for n := 1; n < 7; n++ {
											switch m := ttt.Hours(); {
											case int(m) > 24*n:
												index--
											}
										}
									}

									left = index * p.multiplyer
								case "month":
									index := 30
									t = time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 24, 0, 0, 0, time.Local)
									tt := t.Sub(p.demandRequests[strconv.Itoa(p.filteredRequests[i])].CreatedAt)
									if tt.Hours() > 30*24 || tt.Hours() < 0 {
										p.filteredandValidRequests--
										return nil
									}

									if p.filteredandValidRequests < 2 {
										return nil
									}

									p.validRequests = append(p.validRequests, p.filteredRequests[i])
									if p.filteredRequests[i] == p.validRequests[0] {
										for n := 1; n < 30; n++ {
											switch h := tt.Hours(); {
											case int(h) > 24*n:
												index--
											}
										}
										p.counterSameTime++
										if p.demandRequests[strconv.Itoa(p.filteredRequests[i])].Fulfilled {
											p.counterSameTimeFulfilled++
										}

										if p.demandRequests[strconv.Itoa(p.filteredRequests[i+1])].CreatedAt.Sub(p.demandRequests[strconv.Itoa(p.filteredRequests[i])].CreatedAt).Hours() < 24 {
											return nil
										}
									} else {
										if p.demandRequests[strconv.Itoa(p.filteredRequests[i])].CreatedAt.Sub(p.demandRequests[strconv.Itoa(p.filteredRequests[i-1])].CreatedAt).Hours() < 24 {
											p.counterSameTime++
											if p.demandRequests[strconv.Itoa(p.filteredRequests[i])].Fulfilled {
												p.counterSameTimeFulfilled++
											}
											p.hasChildren = true
											return nil
										}

										var ttt time.Duration

										if p.hasChildren {
											ttt = t.Sub(p.demandRequests[strconv.Itoa(p.filteredRequests[i-1])].CreatedAt)
										} else {
											ttt = tt
											if p.demandRequests[strconv.Itoa(p.filteredRequests[i])].Fulfilled {
												p.counterSameTimeFulfilled++
											}
										}

										for n := 1; n < 30; n++ {
											switch m := ttt.Hours(); {
											case int(m) > 24*n:
												index--
											}
										}
									}

									left = index * p.multiplyer
								case "year":
									index := 12
									t = time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 24, 0, 0, 0, time.Local)
									tt := t.Sub(p.demandRequests[strconv.Itoa(p.filteredRequests[i])].CreatedAt)
									if tt.Hours() > 356*24 || tt.Hours() < 0 {
										p.filteredandValidRequests--
										return nil
									}

									if p.filteredandValidRequests < 2 {
										return nil
									}

									p.validRequests = append(p.validRequests, p.filteredRequests[i])
									if p.filteredRequests[i] == p.validRequests[0] {
										for n := 1; n < 12; n++ {
											switch h := tt.Hours(); {
											case int(h) > 30*24*n:
												index--
											}
										}
										p.counterSameTime++
										if p.demandRequests[strconv.Itoa(p.filteredRequests[i])].Fulfilled {
											p.counterSameTimeFulfilled++
										}

										if p.demandRequests[strconv.Itoa(p.filteredRequests[i+1])].CreatedAt.Sub(p.demandRequests[strconv.Itoa(p.filteredRequests[i])].CreatedAt).Hours() < 30*24 {
											return nil
										}
									} else {
										if p.demandRequests[strconv.Itoa(p.filteredRequests[i])].CreatedAt.Sub(p.demandRequests[strconv.Itoa(p.filteredRequests[i-1])].CreatedAt).Hours() < 30*24 {
											p.counterSameTime++
											if p.demandRequests[strconv.Itoa(p.filteredRequests[i])].Fulfilled {
												p.counterSameTimeFulfilled++
											}
											p.hasChildren = true
											return nil
										}

										var ttt time.Duration

										if p.hasChildren {
											ttt = t.Sub(p.demandRequests[strconv.Itoa(p.filteredRequests[i-1])].CreatedAt)
										} else {
											ttt = tt
											if p.demandRequests[strconv.Itoa(p.filteredRequests[i])].Fulfilled {
												p.counterSameTimeFulfilled++
											}
										}

										for n := 1; n < 12; n++ {
											switch m := ttt.Hours(); {
											case int(m) > 30*24*n:
												index--
											}
										}
									}

									left = index * p.multiplyer
								}

								ratio := (float64(p.counterSameTimeFulfilled) / float64(p.counterSameTime)) * 10

								if t.Sub(p.demandRequests[strconv.Itoa(p.filteredRequests[i])].CreatedAt).Minutes() <= 10 {
									if ratio/10 < 0.6 {
										p.globalEvent = true
										p.resource = strings.ToLower(p.demandRequests[strconv.Itoa(p.filteredRequests[i])].Category)
									}
								}
								top := 390 - int(ratio)*40
								p.coordinates = append(p.coordinates, coordinate{id: i, top: top, left: left})
								if len(p.coordinates) == 1 {
									requests := strconv.Itoa(p.counterSameTime)
									p.counterSameTime = 0
									p.counterSameTimeFulfilled = 0
									p.counterSameTime++
									if p.demandRequests[strconv.Itoa(p.filteredRequests[i])].Fulfilled {
										p.counterSameTimeFulfilled++
									}

									return app.Div().ID("ball"+strconv.Itoa(p.filteredRequests[i])).Class("ball").Style("top", strconv.Itoa(top)+"px").Style("left", strconv.Itoa(left)+"px").Body(
										app.A().Href("#").Body(app.Small().Text(requests+" Requests")),
										app.Div().ID("pulse"+strconv.Itoa(p.filteredRequests[i])).Class("pulse").Style("top", "-1px").Style("left", "-1px"),
									)
								} else {
									last := false
									switch id := p.demandRequests[strconv.Itoa(p.filteredRequests[i])].ID; {
									case id == p.lastWaterRequest:
										last = true
									case id == p.lastFoodRequest:
										last = true
									case id == p.lastHousingRequest:
										last = true
									case id == p.lastOtherRequest:
										last = true
									default:
										last = false
									}

									if last && !p.hasChildren {
										p.counterSameTime++
										if p.demandRequests[strconv.Itoa(p.filteredRequests[i])].Fulfilled {
											p.counterSameTimeFulfilled++
										}
									}

									ratio := (float64(p.counterSameTimeFulfilled) / float64(p.counterSameTime)) * 10
									top := 390 - int(ratio)*40
									prevElement := p.coordinates[len(p.coordinates)-2]
									topDiff := float64(top) - float64(prevElement.top)
									leftDiff := float64(left) - float64(prevElement.left)
									p.coordinates[len(p.coordinates)-1].distance = math.Sqrt(topDiff*topDiff + leftDiff*leftDiff)
									p.coordinates[len(p.coordinates)-1].angle = math.Atan2(float64(prevElement.top)-float64(top), float64(prevElement.left)-float64(left)) * (180 / math.Pi)
									requests := strconv.Itoa(p.counterSameTime)
									p.counterSameTime = 0
									p.counterSameTimeFulfilled = 0
									if !last {
										p.counterSameTime++
										if p.demandRequests[strconv.Itoa(p.filteredRequests[i])].Fulfilled {
											p.counterSameTimeFulfilled++
										}
									}

									return app.Div().ID("ball"+strconv.Itoa(p.filteredRequests[i])).Class("ball").Style("top", strconv.Itoa(top)+"px").Style("left", strconv.Itoa(left)+"px").Body(
										app.A().Href("#").Body(app.Small().Text(requests+" Requests")),
										app.Div().ID("pulse"+strconv.Itoa(p.filteredRequests[i])).Class("pulse").Style("top", "-1px").Style("left", "-1px"),
										app.Div().ID("line"+strconv.Itoa(p.filteredRequests[i])).Class("line").Style("-webkit-transform", "rotate("+fmt.Sprintf("%.2f", p.coordinates[len(p.coordinates)-1].angle)+"deg)").Style("-webkit-transform-origin", "0 0.25em").Style("-webkit-animation", "ball 1s linear forwards").Style("width", fmt.Sprintf("%.2f", p.coordinates[len(p.coordinates)-1].distance)+"px"),
									)
								}
							}),
						),
					),
				),
			),
		),
	)
}

func (p *pubsub) createNotification(ctx app.Context, s NotificationStatus, h, msg string) {
	p.notificationID++
	p.notifications[strconv.Itoa(p.notificationID)] = notification{
		id:      p.notificationID,
		status:  string(s),
		header:  h,
		message: msg,
	}

	ntfs := p.notifications
	ctx.Async(func() {
		for n := range ntfs {
			time.Sleep(5 * time.Second)
			delete(ntfs, n)
			ctx.Async(func() {
				ctx.Dispatch(func(ctx app.Context) {
					p.notifications = ntfs
				})
			})
		}
	})
}

func (p *pubsub) onSelectPeriod(ctx app.Context, e app.Event) {
	p.showRatio = true
	p.showTime = true
	p.showChart = true
	p.validRequests = []int{}
	p.filteredandValidRequests = len(p.filteredRequests)
	p.hasChildren = false
	p.period = ctx.JSSrc().Get("value").String()
	p.setTimeAxis(p.period)
	if p.showRanks {
		app.Window().Get("document").Call("querySelector", "#ranks").Get("classList").Call("remove", "active")
		app.Window().Get("document").Call("querySelector", "#global-stats").Get("classList").Call("add", "active")
		app.Window().Get("document").Call("querySelector", "#category-all").Get("classList").Call("add", "active")
		p.showRanks = false
	} else {
		// remove default period active
		app.Window().Get("document").Call("querySelector", ".period.active").Get("classList").Call("remove", "active")
	}
	// set current period active
	ctx.JSSrc().Get("classList").Call("add", "active")
	// chart rendering
	app.Window().Get("document").Call("querySelector", ".content").Get("classList").Call("remove", "running")
	ctx.Async(func() {
		// wait for rendering
		time.Sleep(1 * time.Millisecond)
		ctx.Dispatch(func(ctx app.Context) {
			// reset counters before re-render
			p.resetChartDefaults()
			// triger chart rendering
			app.Window().Get("document").Call("querySelector", ".content").Get("classList").Call("add", "running")
		})
	})
}

func (p *pubsub) onSelectCategory(ctx app.Context, e app.Event) {
	p.showRatio = true
	p.showTime = true
	p.showChart = true
	p.filteredRequests = []int{}
	p.validRequests = []int{}
	p.filteredandValidRequests = len(p.filteredRequests)
	p.hasChildren = false
	p.category = ctx.JSSrc().Get("value").String()
	switch p.category {
	case "All":
		p.filteredRequests = p.index
	case "Water":
		p.filteredRequests = p.filteredWaterRequests
	case "Food":
		p.filteredRequests = p.filteredFoodRequests
	case "Housing":
		p.filteredRequests = p.filteredHousingRequests
	case "Other":
		p.filteredRequests = p.filteredOtherRequests
	}
	p.filteredandValidRequests = len(p.filteredRequests)
	if p.showRanks {
		app.Window().Get("document").Call("querySelector", "#ranks").Get("classList").Call("remove", "active")
		app.Window().Get("document").Call("querySelector", "#global-stats").Get("classList").Call("add", "active")
		app.Window().Get("document").Call("querySelector", "#period-hour").Get("classList").Call("add", "active")
		p.showRanks = false
		p.setTimeAxis(Hour)
	} else {
		// remove default category active
		app.Window().Get("document").Call("querySelector", ".category.active").Get("classList").Call("remove", "active")
	}
	// set current category active
	ctx.JSSrc().Get("classList").Call("add", "active")
	// chart rendering
	app.Window().Get("document").Call("querySelector", ".content").Get("classList").Call("remove", "running")
	ctx.Async(func() {
		// wait for rendering
		time.Sleep(1 * time.Millisecond)
		ctx.Dispatch(func(ctx app.Context) {
			// reset counters before re-render
			p.resetChartDefaults()
			// triger chart rendering
			app.Window().Get("document").Call("querySelector", ".content").Get("classList").Call("add", "running")
		})
	})
}

func (p *pubsub) onSelectStats(ctx app.Context, e app.Event) {
	p.showRatio = true
	p.showTime = true
	p.showChart = true
	p.stats = ctx.JSSrc().Get("value").String()
	if p.showRanks {
		app.Window().Get("document").Call("querySelector", "#ranks").Get("classList").Call("remove", "active")
		app.Window().Get("document").Call("querySelector", "#category-all").Get("classList").Call("add", "active")
		app.Window().Get("document").Call("querySelector", "#period-hour").Get("classList").Call("add", "active")
		p.showRanks = false
		p.setTimeAxis(Hour)
	} else {
		// remove default stats active
		app.Window().Get("document").Call("querySelector", ".stats.active").Get("classList").Call("remove", "active")
	}
	// set current stats active
	ctx.JSSrc().Get("classList").Call("add", "active")
	// chart rendering
	app.Window().Get("document").Call("querySelector", ".content").Get("classList").Call("remove", "running")
	ctx.Async(func() {
		// wait for rendering
		time.Sleep(1 * time.Millisecond)
		ctx.Dispatch(func(ctx app.Context) {
			// reset counters before re-render
			p.resetChartDefaults()
			// triger chart rendering
			app.Window().Get("document").Call("querySelector", ".content").Get("classList").Call("add", "running")
		})
	})
}

func (p *pubsub) onSelectRanks(ctx app.Context, e app.Event) {
	elems := app.Window().Get("document").Call("querySelectorAll", ".active")
	for i := 0; i < elems.Length(); i++ {
		elems.Index(i).Get("classList").Call("remove", "active")
	}

	app.Window().Get("document").Call("querySelector", ".content").Get("classList").Call("remove", "running")
	// set current category active
	ctx.JSSrc().Get("classList").Call("add", "active")
	p.showRatio = false
	p.showTime = false
	p.showChart = false
	p.showRanks = true
}

func (p *pubsub) resetChartDefaults() {
	p.coordinates = []coordinate{}
	p.counterSameTime = 0
	p.counterSameTimeFulfilled = 0
}

func (p *pubsub) onSelect(ctx app.Context, e app.Event) {
	m := ctx.JSSrc().Get("value").String()
	if m != "" {
		p.demandRequest.Category = m
		if p.demandRequest.Details != "" && p.demandRequest.Quantity != "" {
			enableButton()
		} else {
			disableButton()
		}
	} else {
		disableButton()
	}
	p.resetChartDefaults()
}

func (p *pubsub) onInput(ctx app.Context, e app.Event) {
	m := ctx.JSSrc().Get("value").String()

	if m != "" {
		p.demandRequest.Quantity = m
		if p.demandRequest.Details != "" && p.demandRequest.Category != "" {
			enableButton()
		} else {
			disableButton()
		}
	} else {
		disableButton()
	}
	p.resetChartDefaults()
}

func (p *pubsub) onMessage(ctx app.Context, e app.Event) {
	m := ctx.JSSrc().Get("value").String()

	if m != "" {
		p.demandRequest.Details = m
		if p.demandRequest.Category != "" && p.demandRequest.Quantity != "" {
			enableButton()
		} else {
			disableButton()
		}
	} else {
		disableButton()
	}
	p.resetChartDefaults()
}

func (p *pubsub) sendDemand(ctx app.Context, e app.Event) {
	// Publish to the `topic` through IPFS.
	//
	ctx.Async(func() {
		log.Println("Publisher is about to begin...")
		p.demandRequest.CitizenID = p.citizenID
		p.demandRequest.Fulfilled = false
		p.demandRequest.CreatedAt = time.Now()
		demand, err := json.Marshal(p.demandRequest)
		if err != nil {
			log.Fatal(err)
		}
		// store in orbit-db first
		err = p.sh.OrbitKVPut(dbAddressSupplyDemand, strconv.Itoa(p.demandRequest.ID), []byte(demand))
		if err != nil {
			log.Fatal(err)
		}

		err = p.sh.PubSubPublish(p.topic, string(demand))
		if err != nil {
			log.Fatal(err)
		}

		log.Println("Finished publishing.")
		p.createNotification(ctx, NotificationSuccess, "Demand sent!", "You have requested "+p.demandRequest.Quantity+" "+p.demandRequest.Details+" of "+p.demandRequest.Category+".")
		ctx.Dispatch(func(ctx app.Context) {
			p.showRatio = true
			p.showTime = true
			p.showChart = true
			p.showRanks = false
			p.demandRequests[strconv.Itoa(p.demandRequest.ID)] = p.demandRequest
			p.demandRequest.ID++
		})
	})
}

func (p *pubsub) sendSupply(ctx app.Context, e app.Event) {
	//
	// Publish to the `topic` through IPFS.
	//

	id := ctx.JSSrc().Get("id").String()
	d := p.demandRequests[id]
	d.Fulfilled = true
	d.FulfilledBy = p.citizenID
	d.FulfilledAt = time.Now()

	supply, err := json.Marshal(d)
	if err != nil {
		log.Fatal(err)
	}
	ctx.Async(func() {
		// store in orbit-db first
		err = p.sh.OrbitKVPut(dbAddressSupplyDemand, strconv.Itoa(d.ID), supply)
		if err != nil {
			log.Fatal(err)
		}

		err = p.sh.PubSubPublish(p.topic, string(supply))
		if err != nil {
			log.Fatal(err)
		}
		ctx.Dispatch(func(ctx app.Context) {
			p.demandRequests[id] = d
			p.createNotification(ctx, NotificationSuccess, "Supply sent!", "You have supplied "+d.Quantity+" "+d.Details+" of "+d.Category+".")
		})
	})
}

func (p *pubsub) checkUnsuppliedMessages(ctx app.Context) {
	p.showMessages = false
	if len(p.demandRequests) > 0 {
		for i := range p.index {
			i = i + 1
			for range p.demandRequests {
				if !p.demandRequests[strconv.Itoa(i)].Fulfilled {
					p.showMessages = true
				}
			}
		}
	}
}

func (p *pubsub) FetchAllRequests(ctx app.Context, e app.Event) {
	ctx.Async(func() {
		// query orbit-db
		v, err := p.sh.OrbitKVGet(dbAddressSupplyDemand, "all")
		if err != nil {
			log.Fatal(err)
		}

		ds := make(map[int]string)
		drs := p.demandRequests
		err = json.Unmarshal(v, &ds)
		if err != nil {
			log.Fatal(err)
		}

		p.index = make([]int, 0, len(ds))
		if len(ds) > 0 {
			for i, d := range ds {
				p.index = append(p.index, i)

				dec, err := base64.URLEncoding.DecodeString(string(d))
				if err != nil {
					log.Fatal(err)
				}

				err = json.Unmarshal(dec, &p.demandRequest)
				if err != nil {
					log.Fatal(err)
				}
				if !p.demandRequest.Fulfilled {
					p.showMessages = true
				}

				drs[strconv.Itoa(p.demandRequest.ID)] = p.demandRequest
				if err != nil {
					log.Fatal(err)
				}
			}
		}

		sort.Ints(p.index)
		var nextIndex int
		if len(p.demandRequests) > 0 {
			peerRanks := make([]ranking, 0)
			citizenIDs := make(map[string]bool, 0)
			var peerExists bool
			myDemands := make(map[string]int, 0)
			mySupplies := make(map[string]int, 0)
			totalDemands := 0
			totalSupplies := 0
			r := ranking{}

			for k, v := range p.index {
				if p.citizenID == p.demandRequests[strconv.Itoa(v)].CitizenID {
					p.newComer = false
				}

				if p.demandRequests[strconv.Itoa(v)].CitizenID != p.demandRequests[strconv.Itoa(v-1)].CitizenID {
					totalDemands++
					if !citizenIDs[p.demandRequests[strconv.Itoa(v)].CitizenID] {
						myDemands[p.demandRequests[strconv.Itoa(v)].CitizenID] = 1
						citizenIDs[p.demandRequests[strconv.Itoa(v)].CitizenID] = true
					} else {
						myDemands[p.demandRequests[strconv.Itoa(v)].CitizenID]++
						for kk, vv := range peerRanks {
							if vv.citizenID == p.demandRequests[strconv.Itoa(v)].CitizenID {
								vv.demandRatio = float64(myDemands[p.demandRequests[strconv.Itoa(v)].CitizenID])
								peerRanks[kk].demandRatio = vv.demandRatio
							}
						}
					}
					peerRanks = append(peerRanks, r)
				} else {
					myDemands[p.demandRequests[strconv.Itoa(v)].CitizenID]++
					totalDemands++
				}

				if p.demandRequests[strconv.Itoa(v)].Fulfilled {
					mySupplies[p.demandRequests[strconv.Itoa(v)].FulfilledBy]++
					totalSupplies++
					for kk, vv := range peerRanks {
						if vv.citizenID == p.demandRequests[strconv.Itoa(v)].FulfilledBy {
							vv.supplyRatio = float64(mySupplies[p.demandRequests[strconv.Itoa(v)].FulfilledBy])
							peerRanks[kk].supplyRatio = vv.supplyRatio
							peerExists = true
						}
					}

					if !peerExists {
						r = ranking{
							citizenID:       p.demandRequests[strconv.Itoa(v)].FulfilledBy,
							demandRatio:     0,
							supplyRatio:     float64(mySupplies[p.demandRequests[strconv.Itoa(v)].FulfilledBy]),
							reputationIndex: 0,
						}
						peerRanks = append(peerRanks, r)
					}
				}

				demandRatio := float64(myDemands[p.demandRequests[strconv.Itoa(v)].CitizenID])
				supplyRatio := float64(mySupplies[p.demandRequests[strconv.Itoa(v)].CitizenID])

				r = ranking{
					citizenID:       p.demandRequests[strconv.Itoa(v)].CitizenID,
					demandRatio:     demandRatio,
					supplyRatio:     supplyRatio,
					reputationIndex: 0,
				}

				// next index +1
				nextIndex = p.demandRequests[strconv.Itoa(v)].ID + 1

				switch p.demandRequests[strconv.Itoa(v)].Category {
				case "Water":
					p.lastWaterRequest = p.demandRequests[strconv.Itoa(v)].ID
					p.filteredWaterRequests = append(p.filteredWaterRequests, p.demandRequests[strconv.Itoa(k)].ID)
				case "Food":
					p.lastFoodRequest = p.demandRequests[strconv.Itoa(v)].ID
					p.filteredFoodRequests = append(p.filteredFoodRequests, p.demandRequests[strconv.Itoa(k)].ID)
				case "Housing":
					p.lastHousingRequest = p.demandRequests[strconv.Itoa(v)].ID
					p.filteredHousingRequests = append(p.filteredHousingRequests, p.demandRequests[strconv.Itoa(k)].ID)
				case "Other":
					p.lastOtherRequest = p.demandRequests[strconv.Itoa(v)].ID
					p.filteredOtherRequests = append(p.filteredOtherRequests, p.demandRequests[strconv.Itoa(k)].ID)
				}
			}

			p.filteredRequests = p.index
			p.showChart = true
			p.filteredandValidRequests = len(p.filteredRequests)
			// send welcome notification to newcomers
			if p.newComer {
				p.createNotification(ctx, NotificationPrimary, "Welcome to Cyber Stasis!", "Read How to Play to learn the basics. Please note the game is not optimized for mobile devices. For best experience play it on a computer.")
			}

			ctx.Dispatch(func(ctx app.Context) {
				if len(peerRanks) > 0 {
					for i := range peerRanks {
						if totalDemands > 0 {
							peerRanks[i].demandRatio = peerRanks[i].demandRatio / float64(totalDemands)
						} else {
							peerRanks[i].demandRatio = 0
						}
						if totalSupplies > 0 {
							peerRanks[i].supplyRatio = peerRanks[i].supplyRatio / float64(totalSupplies)
						} else {
							peerRanks[i].supplyRatio = 0
						}
						peerRanks[i].reputationIndex = (peerRanks[i].supplyRatio - peerRanks[i].demandRatio) * ((float64(myDemands[peerRanks[i].citizenID]) + float64(mySupplies[peerRanks[i].citizenID])) / float64(len(p.demandRequests)))
						p.ranks = append(p.ranks, peerRanks[i])
					}
					sort.SliceStable(p.ranks, func(i, j int) bool {
						return p.ranks[i].reputationIndex > p.ranks[j].reputationIndex
					})
				}
				p.demandRequest.ID = nextIndex
				p.demandRequests = drs
				p.resetChartDefaults()
			})
		} else {
			nextIndex = 1
			ctx.Dispatch(func(ctx app.Context) {
				p.demandRequest.ID = nextIndex
			})
		}
	})
}

func (p *pubsub) updateRanks(ctx app.Context) {
	p.ranks = make([]ranking, 0)
	peerRanks := make([]ranking, 0)
	citizenIDs := make(map[string]bool, 0)
	var peerExists bool
	myDemands := make(map[string]int, 0)
	mySupplies := make(map[string]int, 0)
	totalDemands := 0
	totalSupplies := 0
	r := ranking{}
	for _, v := range p.index {
		if p.citizenID == p.demandRequests[strconv.Itoa(v)].CitizenID {
			p.newComer = false
		}

		if p.demandRequests[strconv.Itoa(v)].CitizenID != p.demandRequests[strconv.Itoa(v-1)].CitizenID {
			totalDemands++
			if !citizenIDs[p.demandRequests[strconv.Itoa(v)].CitizenID] {
				myDemands[p.demandRequests[strconv.Itoa(v)].CitizenID] = 1
				citizenIDs[p.demandRequests[strconv.Itoa(v)].CitizenID] = true
			} else {
				myDemands[p.demandRequests[strconv.Itoa(v)].CitizenID]++
				for kk, vv := range peerRanks {
					if vv.citizenID == p.demandRequests[strconv.Itoa(v)].CitizenID {
						vv.demandRatio = float64(myDemands[p.demandRequests[strconv.Itoa(v)].CitizenID])
						peerRanks[kk].demandRatio = vv.demandRatio
					}
				}
			}

			peerRanks = append(peerRanks, r)
		} else {
			myDemands[p.demandRequests[strconv.Itoa(v)].CitizenID]++
			totalDemands++
		}

		if p.demandRequests[strconv.Itoa(v)].Fulfilled {
			mySupplies[p.demandRequests[strconv.Itoa(v)].FulfilledBy]++
			totalSupplies++
			for kk, vv := range peerRanks {
				if vv.citizenID == p.demandRequests[strconv.Itoa(v)].FulfilledBy {
					vv.supplyRatio = float64(mySupplies[p.demandRequests[strconv.Itoa(v)].FulfilledBy])
					peerRanks[kk].supplyRatio = vv.supplyRatio
					peerExists = true
				}
			}

			if !peerExists {
				r = ranking{
					citizenID:       p.demandRequests[strconv.Itoa(v)].FulfilledBy,
					demandRatio:     0,
					supplyRatio:     float64(mySupplies[p.demandRequests[strconv.Itoa(v)].FulfilledBy]),
					reputationIndex: 0,
				}
				peerRanks = append(peerRanks, r)
			}
		}

		demandRatio := float64(myDemands[p.demandRequests[strconv.Itoa(v)].CitizenID])
		supplyRatio := float64(mySupplies[p.demandRequests[strconv.Itoa(v)].CitizenID])

		r = ranking{
			citizenID:       p.demandRequests[strconv.Itoa(v)].CitizenID,
			demandRatio:     demandRatio,
			supplyRatio:     supplyRatio,
			reputationIndex: 0,
		}

		if len(p.index) == 1 {
			peerRanks = append(peerRanks, r)
		}
	}

	ctx.Dispatch(func(ctx app.Context) {
		if len(peerRanks) > 0 {
			for i := range peerRanks {
				if totalDemands > 0 {
					peerRanks[i].demandRatio = peerRanks[i].demandRatio / float64(totalDemands)
				} else {
					peerRanks[i].demandRatio = 0
				}
				if totalSupplies > 0 {
					peerRanks[i].supplyRatio = peerRanks[i].supplyRatio / float64(totalSupplies)
				} else {
					peerRanks[i].supplyRatio = 0
				}
				peerRanks[i].reputationIndex = (peerRanks[i].supplyRatio - peerRanks[i].demandRatio) * ((float64(myDemands[peerRanks[i].citizenID]) + float64(mySupplies[peerRanks[i].citizenID])) / float64(len(p.demandRequests)))
				p.ranks = append(p.ranks, peerRanks[i])
				if peerRanks[i].citizenID == p.citizenID && peerRanks[i].reputationIndex < 0 {
					p.createNotification(ctx, NotificationWarning, "You can do better!", "You need to contribute more!")
				}
			}
			sort.SliceStable(p.ranks, func(i, j int) bool {
				return p.ranks[i].reputationIndex > p.ranks[j].reputationIndex
			})
		}
		if len(p.ranks) > 1 {
			if p.ranks[0].citizenID == p.citizenID {
				p.createNotification(ctx, NotificationSuccess, "Well done!", "You are ranked number one!")
			}
		}
	})
}

func (p *pubsub) deleteRequests(ctx app.Context, e app.Event) {
	ctx.Async(func() {
		// query orbit-db
		err := p.sh.OrbitKVDelete(dbAddressSupplyDemand, "all")
		if err != nil {
			log.Fatal(err)
		}
		ctx.Dispatch(func(ctx app.Context) {
			p.index = make([]int, 0)
			p.filteredRequests = make([]int, 0)
			p.demandRequests = map[string]demandRequest{}
			p.ranks = make([]ranking, 0)
			p.showMessages = false
			p.showChart = false
			p.Update()
		})
	})
}

func (p *pubsub) createDummyRequestHour(n, i, f, id int) {
	dr := demandRequest{}
	dr.ID = id
	dr.Quantity = strconv.Itoa(id)

	var tb int
	t := time.Now()
	if i > 0 {
		tb = t.Hour() - 1
	} else {
		tb = t.Hour()
	}
	tt := t.Add(time.Minute - 10)
	d := (10 * time.Minute)
	dr.CreatedAt = time.Date(t.Year(), t.Month(), t.Day(), tb, tt.Round(d).Minute()+(i*10), 0, 0, time.Local)
	switch n {
	case n % 2:
		dr.Category = "Housing"
		dr.Details = "accommodation"
	case n % 3:
		dr.Category = "Food"
		dr.Details = "kg"
	default:
		dr.Category = "Water"
		dr.Details = "litres"
	}

	if f > n {
		dr.Fulfilled = true
		if i > 0 && i < 5 {
			dr.FulfilledBy = strconv.Itoa(i + 1)
		} else {
			dr.FulfilledBy = strconv.Itoa(1)
		}
		dr.FulfilledAt = time.Date(t.Year(), t.Month(), t.Day(), tb, tt.Round(d).Minute()+(i*10), 0, 0, time.Local)
	}

	dr.CitizenID = strconv.Itoa(i)

	demand, err := json.Marshal(dr)
	if err != nil {
		log.Fatal(err)
	}

	time.Sleep(1 * time.Second)

	// store in orbit-db first
	err = p.sh.OrbitKVPut(dbAddressSupplyDemand, strconv.Itoa(id), demand)
	if err != nil {
		log.Fatal(err)
	}

	err = p.sh.PubSubPublish(p.topic, string(demand))
	if err != nil {
		log.Fatal(err)
	}
}

func (p *pubsub) createDummyRequestDay(n, i, f, id int) {
	dr := demandRequest{}
	dr.ID = id
	dr.Quantity = strconv.Itoa(id)
	if i == 0 {
		i = 23
	}
	t := time.Now()
	dr.CreatedAt = time.Date(t.Year(), t.Month(), t.Day(), i, 0, 0, 0, time.Local)
	switch n {
	case n % 2:
		dr.Category = "Food"
		dr.Details = "kg"
	case n % 3:
		dr.Category = "Housing"
		dr.Details = "accommodation"
	default:
		dr.Category = "Water"
		dr.Details = "litres"
	}

	if f > n {
		dr.Fulfilled = true
		if i < 20 {
			dr.FulfilledBy = strconv.Itoa(i + 10)
		} else {
			dr.FulfilledBy = strconv.Itoa(2)
		}
		dr.FulfilledAt = time.Date(t.Year(), t.Month(), t.Day(), i, 0, 0, 0, time.Local)
	}

	dr.CitizenID = strconv.Itoa(i)

	demand, err := json.Marshal(dr)
	if err != nil {
		log.Fatal(err)
	}

	time.Sleep(1 * time.Second)

	// store in orbit-db first
	err = p.sh.OrbitKVPut(dbAddressSupplyDemand, strconv.Itoa(id), demand)
	if err != nil {
		log.Fatal(err)
	}

	err = p.sh.PubSubPublish(p.topic, string(demand))
	if err != nil {
		log.Fatal(err)
	}
}

func (p *pubsub) createDummyRequestWeek(n, i, f, id int) {
	dr := demandRequest{}
	dr.ID = id
	dr.Quantity = strconv.Itoa(id)
	if i == 0 {
		i = 7
	}
	t := time.Now().AddDate(0, 0, -7)
	dr.CreatedAt = time.Date(t.Year(), t.Month(), t.Day()+i, i, 0, 0, 0, time.Local)
	switch n {
	case n % 2:
		dr.Category = "Housing"
		dr.Details = "accommodation"
	case n % 3:
		dr.Category = "Food"
		dr.Details = "kg"
	default:
		dr.Category = "Water"
		dr.Details = "litres"
	}

	if f > n {
		dr.Fulfilled = true
		if i < 7 {
			dr.FulfilledBy = strconv.Itoa(i + 1)
		} else {
			dr.FulfilledBy = strconv.Itoa(1)
		}
		dr.FulfilledAt = time.Date(t.Year(), t.Month(), t.Day()+i, i, 0, 0, 0, time.Local)
	}

	dr.CitizenID = strconv.Itoa(i)

	demand, err := json.Marshal(dr)
	if err != nil {
		log.Fatal(err)
	}

	time.Sleep(1 * time.Second)

	// store in orbit-db first
	err = p.sh.OrbitKVPut(dbAddressSupplyDemand, strconv.Itoa(id), demand)
	if err != nil {
		log.Fatal(err)
	}

	err = p.sh.PubSubPublish(p.topic, string(demand))
	if err != nil {
		log.Fatal(err)
	}
}

func (p *pubsub) createDummyRequestMonth(n, i, f, id int) {
	dr := demandRequest{}
	dr.ID = id
	dr.Quantity = strconv.Itoa(id)
	if i == 0 {
		i = 30
	}
	t := time.Now().AddDate(0, 0, -30)
	dr.CreatedAt = time.Date(t.Year(), t.Month(), t.Day()+i, i, 0, 0, 0, time.Local)
	switch n {
	case n % 2:
		dr.Category = "Housing"
		dr.Details = "accommodation"
	case n % 3:
		dr.Category = "Food"
		dr.Details = "kg"
	default:
		dr.Category = "Water"
		dr.Details = "litres"
	}

	if f > n {
		dr.Fulfilled = true
		if i < 30 {
			dr.FulfilledBy = strconv.Itoa(i + 1)
		} else {
			dr.FulfilledBy = strconv.Itoa(1)
		}
		dr.FulfilledAt = time.Date(t.Year(), t.Month(), t.Day()+i, i, 0, 0, 0, time.Local)
	}

	dr.CitizenID = strconv.Itoa(i)

	demand, err := json.Marshal(dr)
	if err != nil {
		log.Fatal(err)
	}

	time.Sleep(1 * time.Second)

	// store in orbit-db first
	err = p.sh.OrbitKVPut(dbAddressSupplyDemand, strconv.Itoa(id), demand)
	if err != nil {
		log.Fatal(err)
	}

	err = p.sh.PubSubPublish(p.topic, string(demand))
	if err != nil {
		log.Fatal(err)
	}
}

func (p *pubsub) createDummyRequestYear(n, i, f, id int) {
	dr := demandRequest{}
	dr.ID = id
	dr.Quantity = strconv.Itoa(id)
	if i == 0 {
		i = 12
	}
	t := time.Now().AddDate(-1, i, 0)
	dr.CreatedAt = t
	switch n {
	case n % 2:
		dr.Category = "Housing"
		dr.Details = "accommodation"
	case n % 3:
		dr.Category = "Food"
		dr.Details = "kg"
	default:
		dr.Category = "Water"
		dr.Details = "litres"
	}

	if f > n {
		dr.Fulfilled = true
		if i < 12 {
			dr.FulfilledBy = strconv.Itoa(i + 1)
		} else {
			dr.FulfilledBy = strconv.Itoa(1)
		}
		dr.FulfilledAt = t
	}

	dr.CitizenID = strconv.Itoa(i)

	demand, err := json.Marshal(dr)
	if err != nil {
		log.Fatal(err)
	}

	time.Sleep(1 * time.Second)

	// store in orbit-db first
	err = p.sh.OrbitKVPut(dbAddressSupplyDemand, strconv.Itoa(id), demand)
	if err != nil {
		log.Fatal(err)
	}

	err = p.sh.PubSubPublish(p.topic, string(demand))
	if err != nil {
		log.Fatal(err)
	}
}

func (p *pubsub) createDummyRequestCustom(n, i, f, id int) {
	dr := demandRequest{}
	dr.ID = id
	dr.Quantity = strconv.Itoa(id)
	t := time.Now()
	dr.CreatedAt = t
	switch n {
	// case n % 2:
	// 	dr.Category = "Housing"
	// 	dr.Details = "accommodation"
	// case n % 3:
	// 	dr.Category = "Food"
	// 	dr.Details = "kg"
	default:
		dr.Category = "Water"
		dr.Details = "litres"
	}

	dr.CitizenID = p.citizenID

	demand, err := json.Marshal(dr)
	if err != nil {
		log.Fatal(err)
	}

	time.Sleep(1 * time.Second)

	// store in orbit-db first
	err = p.sh.OrbitKVPut(dbAddressSupplyDemand, strconv.Itoa(id), demand)
	if err != nil {
		log.Fatal(err)
	}

	err = p.sh.PubSubPublish(p.topic, string(demand))
	if err != nil {
		log.Fatal(err)
	}
}

func (p *pubsub) dummyData(ctx app.Context, e app.Event) {
	switch p.period {
	case "hour":
		p.dummyDataHour(ctx, e)
	case "day":
		p.dummyDataDay(ctx, e)
	case "week":
		p.dummyDataWeek(ctx, e)
	case "month":
		p.dummyDataMonth(ctx, e)
	case "year":
		p.dummyDataYear(ctx, e)
	case "custom":
		p.dummyDataCustom(ctx, e)
	}
}

func (p *pubsub) dummyDataHour(ctx app.Context, e app.Event) {
	ctx.Async(func() {
		i := 1
		f := 1
		id := 1
		n := 0
		for i < 6 {
			for n := n; n < 3; n++ {
				p.createDummyRequestHour(n, i, f, id)
				id++
			}
			i = i + 2
			f++
		}

		p.createDummyRequestHour(n, 0, f, id)
	})
}

func (p *pubsub) dummyDataDay(ctx app.Context, e app.Event) {
	ctx.Async(func() {
		i := 2
		f := 1
		id := 1
		n := 0
		for i < 24 {
			for n := n; n < 10; n++ {
				p.createDummyRequestDay(n, i, f, id)
				id++
			}
			i += 10
			f++
		}

		p.createDummyRequestDay(n, 0, f, id)
	})
}

func (p *pubsub) dummyDataWeek(ctx app.Context, e app.Event) {
	ctx.Async(func() {
		i := 1
		f := 1
		id := 1
		n := 0
		for i < 7 {
			for n := n; n < 3; n++ {
				p.createDummyRequestWeek(n, i, f, id)
				id++
			}
			i = i + 2
			f++
		}

		p.createDummyRequestWeek(n, 0, f, id)
	})
}

func (p *pubsub) dummyDataMonth(ctx app.Context, e app.Event) {
	ctx.Async(func() {
		i := 1
		f := 1
		id := 1
		n := 0
		for i < 30 {
			for n := n; n < 3; n++ {
				p.createDummyRequestMonth(n, i, f, id)
				id++
			}
			i += 10
			f++
		}

		p.createDummyRequestMonth(n, 0, f, id)
	})
}

func (p *pubsub) dummyDataYear(ctx app.Context, e app.Event) {
	ctx.Async(func() {
		i := 2
		f := 1
		id := 1
		n := 0
		for i < 12 {
			for n := n; n < 3; n++ {
				p.createDummyRequestYear(n, i, f, id)
				id++
			}
			i += 4
			f++
		}

		p.createDummyRequestYear(n, 0, f, id)
	})
}

func (p *pubsub) dummyDataCustom(ctx app.Context, e app.Event) {
	ctx.Async(func() {
		i := 1
		f := 1
		id := 1
		n := 0
		for i < 11 {
			p.createDummyRequestCustom(n, i, f, id)
			id++
			i++
			f++
		}
	})
}

func (p *pubsub) subscribe(ctx app.Context) {
	ctx.Async(func() {
		subscription, err := p.sh.PubSubSubscribe(p.topic)
		if err != nil {
			log.Fatal(err)
		}
		p.sub = subscription
		p.subscription(ctx)
	})
}

func (p *pubsub) subscription(ctx app.Context) {
	ctx.Async(func() {
		defer p.sub.Cancel()
		// wait on pubsub
		res, err := p.sub.Next()
		if err != nil {
			log.Fatal(err)
		}
		// Decode the string data.
		str := string(res.Data)
		ctx.Async(func() {
			p.subscribe(ctx)
		})
		ctx.Dispatch(func(ctx app.Context) {
			d := demandRequest{}
			err := json.Unmarshal([]byte(str), &d)
			if err != nil {
				log.Fatal(err)
			}

			citizenID := res.From[len(res.From)-8:]
			if citizenID.String() != p.citizenID {
				p.demandRequests[strconv.Itoa(d.ID)] = d
			}
			p.index = append(p.index, d.ID)
			if p.demandRequests[strconv.Itoa(d.ID)].CitizenID == p.citizenID {
				if !p.demandRequests[strconv.Itoa(d.ID)].Fulfilled {
					p.counterDemand++
				} else {
					p.counterDemand--
					p.counterSupply++
				}
			} else {
				p.counterDemand = 0
				p.counterSupply = 0
			}

			if p.counterDemand > 1 && p.counterDemand%10 == 0 {
				p.createNotification(ctx, NotificationInfo, "Demanding!", "You have created "+strconv.Itoa(p.counterDemand/10*10)+" demands.")
			}

			if p.counterSupply > 1 && p.counterSupply%10 == 0 {
				p.createNotification(ctx, NotificationInfo, "Contribution hero!", "You have fulfilled "+strconv.Itoa(p.counterSupply/10*10)+" demands.")

			}

			switch p.demandRequests[strconv.Itoa(d.ID)].Category {
			case "Water":
				p.lastWaterRequest = p.demandRequests[strconv.Itoa(d.ID)].ID
				p.filteredWaterRequests = append(p.filteredWaterRequests, p.demandRequests[strconv.Itoa(d.ID)].ID)
			case "Food":
				p.lastFoodRequest = p.demandRequests[strconv.Itoa(d.ID)].ID
				p.filteredFoodRequests = append(p.filteredFoodRequests, p.demandRequests[strconv.Itoa(d.ID)].ID)
			case "Housing":
				p.lastHousingRequest = p.demandRequests[strconv.Itoa(d.ID)].ID
				p.filteredHousingRequests = append(p.filteredHousingRequests, p.demandRequests[strconv.Itoa(d.ID)].ID)
			case "Other":
				p.lastOtherRequest = p.demandRequests[strconv.Itoa(d.ID)].ID
				p.filteredOtherRequests = append(p.filteredOtherRequests, p.demandRequests[strconv.Itoa(d.ID)].ID)
			}

			switch p.category {
			case "All":
				p.filteredRequests = append(p.filteredRequests, d.ID)
			case "Water":
				p.filteredRequests = p.filteredWaterRequests
			case "Food":
				p.filteredRequests = p.filteredFoodRequests
			case "Housing":
				p.filteredRequests = p.filteredHousingRequests
			case "Other":
				p.filteredRequests = p.filteredOtherRequests
			}

			p.filteredandValidRequests = len(p.filteredRequests)
			p.showChart = true
			p.showRanks = false

			p.resetChartDefaults()
			p.updateRanks(ctx)
			if p.globalEvent {
				header := "Global shortage of " + p.resource + "! "
				msg := "Please supply more " + p.resource + "."
				p.createNotification(ctx, NotificationDanger, header, msg)
				p.sh.PubSubPublish(topicCritical, header)
			}

			p.checkUnsuppliedMessages(ctx)
		})
	})
}

// ** DOM Helpers **/

func enableButton() {
	app.Window().GetElementByID("submitDemand").Set("disabled", false)
}

func disableButton() {
	app.Window().GetElementByID("submitDemand").Set("disabled", true)
}
