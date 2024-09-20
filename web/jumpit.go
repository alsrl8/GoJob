package web

import (
	"GoJob/post"
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

	err := navigateToJumpit(ctx)
	if err != nil {
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

	posts, err := readPosts(ctx)
	if err != nil {
		xlog.Logger.Error(err)
		return
	}

	fmt.Printf("%+v\n", (*posts)[0])
	// TODO Jumpit post를 DB에 저장하기. 이미지와 링크, 세부 정보까지 다루기
}

func readPosts(ctx context.Context) (*[]post.JumpitPost, error) {
	var postNodes []*cdp.Node
	var posts []post.JumpitPost

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
		var postName string
		var skillStack string
		var description string

		err = chromedp.Run(ctx,
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
		)
		if err != nil {
			xlog.Logger.Error(err)
			fmt.Printf("Post %d: Failed to retrieve _post details.\n")
			continue
		}

		posts = append(posts, post.JumpitPost{
			Name:        postName,
			Skills:      strings.Split(skillStack, "\n· "),
			Description: description,
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
