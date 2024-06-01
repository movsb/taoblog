//
//  EditorView.swift
//  TaoBlog
//
//  Created by Yang Tao on 2024/5/31.
//

import SwiftUI

struct EditorView: View {
	@EnvironmentObject var states: GlobalState
	
	@Binding var post: Protocols_Post
	
	// 不能直接绑 $post.source，否则光标会乱跑
	@State var source: String
	
	init(post: Binding<Protocols_Post>) {
		self._post = post
		self.source = post.wrappedValue.source
	}
	
	@State private var _alertMessage = ""
	@State private var _alerting = false

	var body: some View {
		TextEditor(text: $source)
			.fontDesign(.monospaced)
		.padding([.leading, .trailing], 4)
		.scrollContentBackground(.hidden)
		.background(Color.gray.opacity(0.1))
		.toolbar {
			ToolbarItem(placement: .topBarTrailing) {
				Button {
					if post.id == 0 {
						post.sourceType = "markdown"
						post.type = "tweet"
						post.source = source
						states.client!.createPost(post) { result in
							switch result {
							case .success(let success):
								post = success
								_alertMessage = "发表成功。"
								_alerting = true
							case .failure(let failure):
								_alertMessage = "发表失败\(failure)。"
								_alerting = true
							}
						}
					} else {
						post.source = source
						states.client!.updatePost(post) { result in
							switch result {
							case .success(let success):
								post = success
								_alertMessage = "更新成功。"
								_alerting = true
							case .failure(let failure):
								_alertMessage = "更新失败\(failure)。"
								_alerting = true
							}
						}
					}
				} label: {
					Text("保存")
				}
			}
		}
		.alert(isPresented: $_alerting, content: {
			Alert(title: Text(_alertMessage))
		})
	}
}

//#Preview {
//	EditorView()
//}
