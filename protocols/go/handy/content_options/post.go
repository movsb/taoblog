package co

import "github.com/movsb/taoblog/protocols/go/proto"

// 这个包存在的目的是为了强制设置需要的参数，而不至于在添加
// 参数后忘记在某些地方同步这些默认参数值。
//
// 当然，写足够多的单测就不怕被破坏、也就不需要这个了。

/*
func ContentOptions(
	withContent bool,
	keepTitleHeading bool,
	renderCodeBlocks bool,
	useAbsolutePaths bool,
	prettifyHtml bool,
	openLinksInNewTab proto.PostContentOptions_OpenLinkInNewTabKind,
) *proto.PostContentOptions {
	return &proto.PostContentOptions{
		WithContent:       withContent,
		KeepTitleHeading:  keepTitleHeading,
		RenderCodeBlocks:  renderCodeBlocks,
		UseAbsolutePaths:  useAbsolutePaths,
		PrettifyHtml:      prettifyHtml,
		OpenLinksInNewTab: openLinksInNewTab,
	}
}
*/

type Kind byte

const (
	Editor Kind = iota + 1
	ClientGetPost
	CreatePost
	GetPostComments
	CheckPostTaskListItems
	GenerateTweetTitle

	CreateCommentCheck
	CreateCommentReturn
	CreateCommentGetPost
	UpdateCommentCheck
	UpdateCommentReturn
	PreviewComment

	SearchIndex

	HomeLatestPosts
	HomeLatestTweets
	HomeLatestComments
	HomeLatestCommentsPosts

	Tweets

	QueryByID
	QueryByPage

	Rss
)

type CO = proto.PostContentOptions

var co = map[Kind]*CO{
	Editor: {
		WithContent: false,
	},
	ClientGetPost: {
		WithContent: false,
	},
	CreatePost: {
		WithContent:       true,
		KeepTitleHeading:  true,
		RenderCodeBlocks:  true,
		OpenLinksInNewTab: proto.PostContentOptions_OpenLinkInNewTabKindAll,
		UseAbsolutePaths:  true,
	},
	GetPostComments: {
		WithContent:       true,
		RenderCodeBlocks:  true,
		UseAbsolutePaths:  false,
		OpenLinksInNewTab: proto.PostContentOptions_OpenLinkInNewTabKindAll,
	},
	CheckPostTaskListItems: {
		WithContent: false,
	},
	GenerateTweetTitle: {
		WithContent:  true,
		PrettifyHtml: true,
	},
	CreateCommentCheck: {
		WithContent: false,
	},
	CreateCommentReturn: {
		WithContent:       true,
		RenderCodeBlocks:  true,
		OpenLinksInNewTab: proto.PostContentOptions_OpenLinkInNewTabKindAll,
	},
	CreateCommentGetPost: {
		WithContent:  true,
		PrettifyHtml: true,
	},
	UpdateCommentCheck: {
		WithContent: false,
	},
	UpdateCommentReturn: {
		WithContent:       true,
		RenderCodeBlocks:  true,
		OpenLinksInNewTab: proto.PostContentOptions_OpenLinkInNewTabKindAll,
	},
	PreviewComment: {
		WithContent:       true,
		RenderCodeBlocks:  true,
		OpenLinksInNewTab: proto.PostContentOptions_OpenLinkInNewTabKindAll,
	},
	SearchIndex: {
		WithContent: false,
	},
	HomeLatestPosts: {
		WithContent:  true,
		PrettifyHtml: true,
	},
	HomeLatestTweets: {
		WithContent:  true,
		PrettifyHtml: true,
	},
	HomeLatestComments: {
		WithContent:  true,
		PrettifyHtml: true,
	},
	HomeLatestCommentsPosts: {
		WithContent: false,
	},
	Tweets: {
		WithContent:       true,
		RenderCodeBlocks:  true,
		UseAbsolutePaths:  true,
		OpenLinksInNewTab: proto.PostContentOptions_OpenLinkInNewTabKindAll,
	},

	QueryByID: {
		WithContent:       true,
		RenderCodeBlocks:  true,
		UseAbsolutePaths:  false,
		OpenLinksInNewTab: proto.PostContentOptions_OpenLinkInNewTabKindAll,
	},
	QueryByPage: {
		WithContent:       true,
		RenderCodeBlocks:  true,
		UseAbsolutePaths:  false,
		OpenLinksInNewTab: proto.PostContentOptions_OpenLinkInNewTabKindAll,
	},

	Rss: {
		WithContent:      true,
		RenderCodeBlocks: false,
	},
}

func For(kind Kind) *CO {
	if o, ok := co[kind]; ok {
		return o
	}
	panic(`无效的内容选项。`)
}
