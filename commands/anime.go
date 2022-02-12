package commands

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"golang.org/x/net/html"
)

var url = "https://v2.vost.pw/"

type Anime struct {
	Title   string
	TitleEn string
	Url     string
	//Description string

}

func getHtml(url string) (string, error) {

	html, err := http.Get(url)

	if err != nil {
		fmt.Println("Error while reciving html request")
	}

	if html.StatusCode != 200 {
		fmt.Println("Bed request status code")
	}

	defer html.Body.Close()

	body, err := ioutil.ReadAll(html.Body)

	if err != nil {

		return "", err
	}

	return string(body), nil

}

func GetAnimeList() *[]Anime {

	data, err := getHtml(url)

	//todo сделать тип для передачи аним
	AnimeList := []Anime{}

	if err != nil {
		fmt.Println("Error while reading html")
	}

	tkn, err := html.Parse(strings.NewReader(data))

	if err != nil {
		fmt.Println("fail")
		return nil
	}

	var f func(*html.Node)

	// перебор всех элементов DOM дерева

	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "div" {
			for _, div := range n.Attr {
				// поиск элемента DOM дерева с типом div и аттрибутом class
				if div.Key == "class" && div.Val == "shortstoryHead" {

					/*
					*
					*  todo сделать нормально
					*
					 */

					for c := n.FirstChild.NextSibling.FirstChild; c != nil; c = c.NextSibling {
						if c.Type == html.ElementNode && c.Data == "a" {

							var title = strings.Split(" / ", c.FirstChild.Data)[0]
							var titleEn = strings.Split(" / ", c.FirstChild.Data)[2]

							AnimeList = append(AnimeList, Anime{title, titleEn, c.Attr[0].Val})

							//fmt.Sprintf("Ну что пацаны, аниме? %s, Ссылка: %s\n", c.FirstChild.Data, c.Attr[0].Val))

						}
					}
					return
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}

	f(tkn)

	return &AnimeList

}
