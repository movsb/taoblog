import { useCallback, useEffect, useRef, useState } from 'react'
import { Tldraw, getSnapshot, useEditor } from 'tldraw'
import 'tldraw/tldraw.css'
import './App.css'

function SnapshotToolbar() {
	const editor = useEditor();

	const save = useCallback(async () => {
		const shapeIds = editor.getCurrentPageShapeIds();
		if(shapeIds.size <= 0) {
			alert('不能保存，因为画布上没有任何内容。');
			return false;
		} 

		const { blob: lightBlob, width, height } = await editor.toImage([...shapeIds], {
			background: false,
			format: 'svg',
			darkMode: false,
			padding: 16,
		});
		const { blob: darkBlob } = await editor.toImage([...shapeIds], {
			background: false,
			format: 'svg',
			darkMode: true,
			padding: 16,
		});

		console.log('image length:', lightBlob.size, width, height);
		// console.log((await blob.bytes()).toBase64());

		const { document, session } = getSnapshot(editor.store);

		/** @type {(state: string, lightBlob: Blob, darkBlob: Blob) => Promise<boolean>} */
		const saver = editor.__saveSnapshot;
		return await saver(
			JSON.stringify({ document, session }),
			lightBlob, darkBlob,
		);
	}, [editor])

	const close = useCallback(async() => {
		const close = editor.__close;
		return await close();
	}, [editor]);

	const [saveState, setSaveState] = useState('idle')
	const saving = useRef(false)
	const showCheckMark = saveState === 'saved'
	const showSaving = saveState === 'saving'
	useEffect(() => {
		if (saveState === 'saved') {
			const timeout = setTimeout(() => {
				setSaveState('idle')
			}, 1000)
			return () => clearTimeout(timeout)
		}
		return
	}, [saveState])

	return (
		<div className='share-buttons'>
			<span
				style={{
					display: 'inline-block',
					transition: 'transform 0.2s ease, opacity 0.2s ease',
					transform: showCheckMark ? `scale(1)` : `scale(0.5)`,
					opacity: showCheckMark ? 1 : 0,
				}}
			>✅ 已保存</span>
			<span
				style={{
					transition: 'transform 0.2s ease, opacity 0.2s ease',
					transform: showSaving ? `scale(1)` : `scale(0.5)`,
					opacity: showSaving ? 1 : 0,
					display: showSaving ? 'inline-block' : 'none',
				}}
			>保存中...</span>
			<button type='button' disabled={showSaving}
				onClick={async () => {
					if (saving.current) { return; }
					saving.current = true;
					setSaveState('saving');
					let saved = false;
					try {
						saved = await save();
					} catch (error) {
						alert('保存失败：' + (error?.message ?? error));
					} finally {
						saving.current = false;
						setSaveState(saved ? 'saved' : 'idle');
					}
				}}
			>保存</button>
			<button type='button' onClick={async () => {
				await close();
			}}>关闭</button>
		</div>
	)
}

export default function App({snapshotJSON, saveSnapshot, close}) {
	return (
		<div className='tldraw-app'>
			<Tldraw
				inferDarkMode
				snapshot={JSON.parse(snapshotJSON ?? '{}')}
				onMount={ed => {
					ed.__saveSnapshot = saveSnapshot;
					ed.__close = close;
				}}
				options={{
					maxPages: 1,
				}}
				components={{
					SharePanel: SnapshotToolbar,
				}}
			/>
		</div>
	)
}
