#pragma once

namespace taoblog {

using Callable = std::function<ComRet(DispParamsVisitor args, VARIANT* result)>;

class WebBrowserVersionSetter
{
public:
    WebBrowserVersionSetter();
    ~WebBrowserVersionSetter();

protected:
    int GetMSHTMLVersion();
    void InitInternetExplorer();
    void UnInitInternetExplorer();

protected:
    std::wstring _name;
};

class EventDelegate
{
public:
    virtual void OnFocusChange(bool focus) {}
    virtual void OnBeforeNavigate(const wchar_t* uri, bool top) {}
    virtual void OnNewWindow(const wchar_t* uri, const wchar_t* ref, bool* pCancel, IDispatch** ppDisp) { *pCancel = false; *ppDisp = nullptr; }
    virtual void OnNavigateComplete(const wchar_t* uri, bool top) {}
    virtual void OnDocumentComplete(const wchar_t* uri, bool top) {}
    virtual void OnTitleChange(const wchar_t* title) {}
    virtual void OnFaviconURLChange(const wchar_t* uri) {}
    virtual void OnSetStatusText(const wchar_t* text) {}
    virtual void OnCommandStatusChange(bool cangoback, bool cangofwd) {}
    virtual void OnLoadError(bool is_main_frame, const wchar_t* url, int error_code, const wchar_t* error_text) {}
    virtual void OnWindowClosing() {}
};

class IWebBrowserContainer
{
public:
    virtual void Create(HWND hParent) = 0;
    virtual void Destroy() = 0;
    virtual void SetDelegate(EventDelegate* pDelegate) = 0;
    virtual void SetPos(const RECT& pos) = 0;
    virtual bool Focus() = 0;
    virtual void SetFocus() = 0;
    virtual bool FilterMessage(MSG* msg) = 0;

    virtual void Navigate(const wchar_t* url) = 0;
    virtual void GoHome() = 0;
    virtual void GoForward() = 0;
    virtual void GoBack() = 0;
    virtual void Refresh(bool force) = 0;
    virtual void Stop() = 0;

    virtual ComRet GetDocument(IHTMLDocument2** ppDocument) = 0;
	virtual std::wstring GetSource() = 0;
    virtual ComRet GetRootElement(IHTMLElement** ppElement) = 0;
    virtual ComRet ExecScript(const std::wstring& script, VARIANT* result, const std::wstring& lang) = 0;

    virtual void AddCallable(const wchar_t* name, Callable call) = 0;
    virtual void RemoveCallable(const wchar_t* name) = 0;
    virtual void FireEvent(const wchar_t* name, UINT argc, VARIANT* argv) = 0;

protected:

};

IWebBrowserContainer* CreateBroserInstance();

} // namespace taoblog
