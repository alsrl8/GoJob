package web

import (
	"GoJob/db"
	"GoJob/info"
	"GoJob/xlog"
	"context"
	"errors"
	"fmt"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
	"github.com/rivo/tview"
	"strconv"
	"strings"
	"time"
)

func CrawlJumpit() {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	err := clearJumpitPostData()
	if err != nil {
		xlog.Logger.Error(err)
		return
	}

	err = navigateToJumpit(ctx)
	if err != nil {
		xlog.Logger.Error(err)
		return
	}

	postNum, err := getJumpitPostNum(ctx)
	if err != nil {
		xlog.Logger.Error(err)
		return
	}
	xlog.Logger.Info(fmt.Sprintf("Golang post number in jumpit: %d", postNum))

	err = scrollPostList(ctx, postNum)
	if err != nil {
		xlog.Logger.Error(err)
		return
	}

	posts, err := readJumpitPosts(ctx)
	if err != nil {
		xlog.Logger.Error(err)
		return
	}

	storeJumpitPosts(posts)
}

func CrawlJumpitPostDetail(index int, detailView *tview.TextView, stopLoading chan bool) {
	sqlite := db.NewSqlite()
	data, err := sqlite.SelectData("jumpit", "")
	if err != nil {
		stopLoading <- true
		xlog.Logger.Error(err)
		return
	}
	link, ok := data[index]["link"].(string)
	if !ok {
		stopLoading <- true
		xlog.Logger.Error("failed to get url")
		return
	}

	url := fmt.Sprintf("https://www.jumpit.co.kr%s", link)

	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	err = navigateToJumpitDetail(ctx, url)
	if err != nil {
		stopLoading <- true
		xlog.Logger.Error(err)
		return
	}

	text, err := readJumpitDetail(ctx)
	if err != nil {
		stopLoading <- true
		xlog.Logger.Error(err)
		return
	}

	stopLoading <- true
	detailView.SetText(text)
}

func readJumpitDetail(ctx context.Context) (string, error) {
	var headerNodes []*cdp.Node
	err := chromedp.Run(ctx,
		chromedp.Nodes(
			`sc-f491c6ef-0 egBfVn`,
			&headerNodes,
			chromedp.NodeVisible,
		),
	)
	if err != nil {
		xlog.Logger.Error(err)
		return "", errors.New("failed to read headerNodes")
	}

	var congratulatoryMoney string
	err = chromedp.Run(ctx,
		chromedp.Text(`div.sc-f491c6ef-1.bManvq > span`,
			&congratulatoryMoney,
			chromedp.NodeVisible,
			chromedp.ByQuery,
		),
	)
	if err != nil {
		xlog.Logger.Error(err)
		return "", errors.New("failed to read congratulatoryMoney")
	}

	return congratulatoryMoney, nil
}

func clearJumpitPostData() error {
	sqlite := db.NewSqlite()
	err := sqlite.DeleteData("jumpit", "")
	return err
}

func storeJumpitPosts(posts *[]*info.JumpitPost) {
	sqlite := db.NewSqlite()

	for _, post := range *posts {
		data := map[string]interface{}{
			"name":        post.Name,
			"description": post.Description,
			"company":     post.Company,
			"skills":      strings.Join(post.Skills, ","),
			"link":        post.Link,
		}
		err := sqlite.InsertData("jumpit", data)
		if err != nil {
			xlog.Logger.Error(err)
			continue
		}
	}
}

func readJumpitPosts(ctx context.Context) (*[]*info.JumpitPost, error) {
	var postNodes []*cdp.Node
	var posts []*info.JumpitPost

	err := chromedp.Run(ctx,
		chromedp.Nodes(
			`.sc-d609d44f-0.grDLmW`,
			&postNodes,
			chromedp.NodeVisible,
		),
	)
	if err != nil {
		xlog.Logger.Error(err)
		return nil, errors.New("failed to read postNodes")
	}

	for _, _post := range postNodes {
		var companyName string
		var postName string
		var skillStack string
		var description string
		var link string

		err = chromedp.Run(ctx,
			chromedp.Text(
				`div.sc-15ba67b8-0.kkQQfR > div > div`,
				&companyName,
				chromedp.FromNode(_post),
				chromedp.NodeVisible,
				chromedp.ByQuery,
			),
			chromedp.Text(
				`:scope .position_card_info_title`,
				&postName,
				chromedp.FromNode(_post),
				chromedp.NodeVisible,
				chromedp.ByQuery,
			),
			chromedp.Text(
				`:scope .sc-15ba67b8-1.iFMgIl`,
				&skillStack,
				chromedp.FromNode(_post),
				chromedp.NodeVisible,
				chromedp.ByQuery,
			),
			chromedp.Text(
				`:scope .sc-15ba67b8-1.cdeuol`,
				&description,
				chromedp.FromNode(_post),
				chromedp.NodeVisible,
				chromedp.ByQuery,
			),
			chromedp.AttributeValue(
				`:scope a`,
				"href",
				&link,
				nil,
				chromedp.FromNode(_post),
				chromedp.NodeVisible,
				chromedp.ByQuery,
			),
		)
		if err != nil {
			xlog.Logger.Error(err)
			continue
		}

		posts = append(posts, &info.JumpitPost{
			Company:     companyName,
			Name:        postName,
			Skills:      strings.Split(skillStack, "\n· "),
			Description: description,
			Link:        link,
		})
	}

	return &posts, nil
}

func scrollPostList(ctx context.Context, postNum int) error {
	var loadedPostNum int
	var loadedPosts []*cdp.Node

	err := chromedp.Run(ctx,
		chromedp.ActionFunc(func(ctx context.Context) error {
			for loadedPostNum < postNum {
				err := chromedp.Run(ctx,
					chromedp.Evaluate(`window.scrollTo(0, document.documentElement.scrollHeight)`, nil),
					chromedp.Sleep(2*time.Second),
					chromedp.Nodes(`.sc-d609d44f-0.grDLmW`, &loadedPosts, chromedp.NodeVisible),
				)
				if err != nil {
					xlog.Logger.Error(err)
					return errors.New("failed to scroll post list")
				}

				loadedPostNum = len(loadedPosts)
			}
			return nil
		}),
	)

	if err != nil {
		xlog.Logger.Error(err)
		return errors.New("failed to scroll post list")
	}

	return nil
}

func getJumpitPostNum(ctx context.Context) (int, error) {
	var postNumStr string
	err := chromedp.Run(ctx,
		chromedp.Sleep(1*time.Second),
		chromedp.Text(".sc-c12e57e5-2.fFIvdj", &postNumStr, chromedp.ByQuery),
	)

	if err != nil {
		xlog.Logger.Error(err)
		return 0, errors.New("failed to get post number")
	}

	sp := strings.Split(postNumStr, "(")
	postNum, err := strconv.Atoi(sp[1][:len(sp[1])-1])
	if err != nil {
		xlog.Logger.Error(err)
		return 0, errors.New("failed to get post number")
	}

	return postNum, nil
}

func navigateToJumpit(ctx context.Context) error {
	url := "https://www.jumpit.co.kr/search?sort=relation&keyword=golang"

	err := chromedp.Run(ctx, chromedp.Navigate(url))
	if err != nil {
		xlog.Logger.Error(err)
		return errors.New("failed to navigate to jumpit")
	}

	return nil
}

func navigateToJumpitDetail(ctx context.Context, url string) error {
	err := chromedp.Run(ctx, chromedp.Navigate(url))
	if err != nil {
		xlog.Logger.Error(err)
		return errors.New("failed to navigate to jumpit detail")
	}

	return nil
}
