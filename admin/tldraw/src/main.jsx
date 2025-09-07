import { createRoot } from 'react-dom/client'
import App from './App.jsx'

// 给纯 JS 使用的渲染函数
window.renderTldrawComponent = function (placeholder, props) {
  const root = createRoot(placeholder);
  root.render(<App {...props} />);
};
