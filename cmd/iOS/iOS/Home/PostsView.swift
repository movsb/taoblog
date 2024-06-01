//
//  PostsView.swift
//  TaoBlog
//
//  Created by Yang Tao on 2024/6/1.
//

import SwiftUI

struct PostView : View {
	@EnvironmentObject var states: GlobalState
	@Environment(\.editMode) private var editMode
	@Binding var p: Protocols_Post
	
	public init(_ p: Binding<Protocols_Post>) {
		self._p = p
	}
	
	var body: some View {
		ZStack {
			if editMode?.wrappedValue.isEditing == true {
				EditorView(post: $p)
					.autocorrectionDisabled(true)
					.navigationTitle("编辑文章")
					.navigationBarTitleDisplayMode(.inline)
			} else {
				WebView(url: "\(states.url)/\($p.id)")
					.navigationTitle("预览文章")
					.navigationBarTitleDisplayMode(.inline)
			}
		}
		.toolbar {
			ToolbarItem(placement: .topBarTrailing) {
				EditButton()
			}
		}
	}
}

struct PostsView: View {
	@EnvironmentObject var states: GlobalState
	
	@State private var _alertMessage = ""
	@State private var _alerting = false
	
	@State private var _posts = [Protocols_Post]()

	@State private var _showNewPost = false
	@State private var _newPost = Protocols_Post()
	
	var body: some View {
		NavigationView {
			List {
				ForEach($_posts) { $p in
					NavigationLink {
						PostView($p)
					} label: {
						Image(systemName: p.status == "public" ? "" : "lock")
							.frame(width: 18, height: 18)
						Text(p.title)
							.lineLimit(1)
							.swipeActions(edge: .trailing, allowsFullSwipe: false) {
								Button(p.status == "public" ? "私密" : "公开") {
									let newStatus = p.status == "public" ? "draft" : "public"
									try! states.client?.setPostStatus(p.id, status: newStatus) {
										p.status = newStatus
									}
								}
								.tint(.accentColor)
							}
					}
				}
			}
			.listStyle(.plain)
			.navigationTitle("近期文章")
			.navigationBarTitleDisplayMode(.inline)
			.refreshable {
				try! states.client?.listPosts {
					_posts = $0
				}
			}
			.toolbar {
				Button {
					_newPost = Protocols_Post()
					_showNewPost = true
				} label: {
					Image(systemName: "plus")
				}
				.fullScreenCover(isPresented: $_showNewPost) {
					NavigationView {
						EditorView(post: $_newPost)
							.navigationTitle(_newPost.id > 0 ? "编辑文章" : "发表文章")
							.navigationBarTitleDisplayMode(.inline)
							.toolbar {
								ToolbarItem(placement: .topBarLeading) {
									Button {
										_showNewPost = false
										if _newPost.id > 0 {
											_posts.insert(_newPost, at: 0)
										}
									} label: {
										Text("关闭")
									}
								}
							}
					}
				}
			}
		}
		.onAppear {
			try! states.client?.listPosts {
				_posts = $0
			}
		}
	}
}

#Preview {
    PostsView()
}
