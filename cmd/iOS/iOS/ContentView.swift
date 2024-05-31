//
//  ContentView.swift
//  TaoBlog
//
//  Created by Yang Tao on 2024/5/31.
//

import SwiftUI
import SwiftDown

extension Protocols_Post : Identifiable { }

struct PostView : View {
	@State private var _post: Protocols_Post
	
	init(_ _post: Protocols_Post) {
		self._post = _post
	}
	
	var body: some View {
		SwiftDownEditor(text: $_post.source)
			.theme(.BuiltIn.defaultLight.theme())
	}
}

struct ContentView: View {
	@State private var _client = Client()
	@State private var _posts = [Protocols_Post]()
	
	var body: some View {
		VStack {
			NavigationView {
				Group {
					List {
						ForEach(_posts) { _post in
							NavigationLink {
								PostView(_post)
							} label: {
								Text(_post.title)
							}
						}
					}
				}
			}
		}
		.onAppear {
			try? self._client.listPosts {
				_posts = $0
			}
		}
	}
}

#Preview {
    ContentView()
}
