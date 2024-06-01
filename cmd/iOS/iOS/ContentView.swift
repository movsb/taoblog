//
//  ContentView.swift
//  TaoBlog
//
//  Created by Yang Tao on 2024/5/31.
//

import SwiftUI

extension Protocols_Post : Identifiable { }

struct ContentView: View {
	@EnvironmentObject var states : GlobalState
	
	var body: some View {
		if states.signedIn {
			TabView {
				HomeView()
					.tabItem {
						Label("首页", systemImage: "house")
					}
				PostsView()
					.tabItem {
						Label("文章", systemImage: "square.and.pencil")
					}
				SettingsView()
					.tabItem {
						Label("设置", systemImage: "gear")
					}
			}
		} else {
			LoginView()
		}
	}
}

#Preview {
    ContentView()
}
