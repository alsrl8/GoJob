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

func CrawlJumpitPostDetail(index int) (dtl *info.JumpitDetail) {
	sqlite := db.NewSqlite()
	data, err := sqlite.SelectData("jumpit", "")
	if err != nil {
		xlog.Logger.Error(err)
		return
	}

	link, ok := data[index]["link"].(string)
	if !ok {
		xlog.Logger.Error("failed to get url")
		return
	}

	url := fmt.Sprintf("https://www.jumpit.co.kr%s", link)

	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err = navigateToJumpitDetail(ctx, url)
	if err != nil {
		xlog.Logger.Error(err)
		return
	}

	dtl, err = readJumpitDetail(ctx)
	if err != nil {
		xlog.Logger.Error(err)
		return
	}

	return
}

func readJumpitDetail(ctx context.Context) (dtl *info.JumpitDetail, err error) {
	var tagContainer []*cdp.Node
	err = chromedp.Run(ctx,
		chromedp.Nodes(
			`//main/div/div/section/div/ul`,
			&tagContainer,
			chromedp.NodeVisible,
		),
	)

	tagContainerNode := tagContainer[0]

	var tagNodes []*cdp.Node
	err = chromedp.Run(ctx,
		chromedp.Nodes(
			tagContainerNode.FullXPathByID()+`/li`,
			&tagNodes,
			chromedp.NodeVisible,
		),
	)
	if err != nil {
		xlog.Logger.Error(err)
		return
	}

	var tags []string
	for _, tag := range tagNodes {
		var tagStr string
		err = chromedp.Run(ctx,
			chromedp.Text(
				tag.FullXPathByID(),
				&tagStr,
				chromedp.NodeVisible,
			),
		)
		if err != nil {
			xlog.Logger.Error(err)
			return
		}
		tags = append(tags, tagStr)
	}

	var congratulations string
	err = chromedp.Run(ctx,
		chromedp.Text(`//main/div/div[1]/section/div/div[1]/span`, &congratulations, chromedp.NodeVisible),
	)
	if err != nil {
		xlog.Logger.Error(err)
		return
	}

	dtl = &info.JumpitDetail{}
	dtl.Congratulations = congratulations
	dtl.Tags = strings.Join(tags, "\n")

	return
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
			`//main/div/section[2]/section/div`,
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
				_post.FullXPathByID()+`/a/div[3]/div`,
				&companyName,
				chromedp.NodeVisible,
			),
			chromedp.Text(
				_post.FullXPathByID()+`/a/div[3]/h2`,
				&postName,
				chromedp.NodeVisible,
			),
			chromedp.Text(
				_post.FullXPathByID()+`/a/div[3]/ul[1]`,
				&skillStack,
				chromedp.FromNode(_post),
				chromedp.NodeVisible,
			),
			chromedp.Text(
				_post.FullXPathByID()+`/a/div[3]/ul[2]`,
				&description,
				chromedp.FromNode(_post),
				chromedp.NodeVisible,
			),
			chromedp.AttributeValue(
				_post.FullXPathByID()+`/a`,
				"href",
				&link,
				nil,
				chromedp.NodeVisible,
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
					chromedp.Nodes(`//main/div/section[2]/section/div`, &loadedPosts, chromedp.NodeVisible),
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
		chromedp.Text("//main/div/section[2]/div/div", &postNumStr, chromedp.NodeVisible),
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
