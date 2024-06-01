//
//  HomeView.swift
//  TaoBlog
//
//  Created by Yang Tao on 2024/5/31.
//

import SwiftUI

struct HomeView: View {
	@EnvironmentObject var states: GlobalState
	@State var info = Protocols_GetInfoResponse()
	
	var body: some View {
		NavigationView{
			WebView(url: states.url)
				.navigationTitle(info.name)
				.navigationBarTitleDisplayMode(.inline)
		}
		.onAppear {
			states.client?.getInfo {
				self.info = $0
			}
		}
	}
}

