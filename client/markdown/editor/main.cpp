#include <string>
#include <functional>
#include <fstream>
#include <cstdio>
#include <string>
#include <atlbase.h>
#include <shobjidl.h>
#include <windows.h>
#include <memory>
#include <codecvt>

#include <atlbase.h>
#include <atlwin.h>

#include "com_hlp.hpp"
#include "webview.h"

namespace taoblog {

#define BIND(c, f) std::bind(&c::f, this, std::placeholders::_1, std::placeholders::_2)

typedef std::wstring_convert<std::codecvt_utf8_utf16<wchar_t>, wchar_t> U8U16Cvt;

std::wstring app_path()
{
    wchar_t path[MAX_PATH];
    ::GetModuleFileName(nullptr, path, _countof(path));
    *wcsrchr(path, L'\\') = L'\0';
    return path;
}

static std::wstring file_open_dialog(HWND hOwner)
{
    std::wstring ret;
    CComPtr<IFileOpenDialog> spFileOpenDialog;
    if(SUCCEEDED(spFileOpenDialog.CoCreateInstance(__uuidof(FileOpenDialog)))) {
        FILEOPENDIALOGOPTIONS options;
        if(SUCCEEDED(spFileOpenDialog->GetOptions(&options))) {
            spFileOpenDialog->SetOptions(options | FOS_FORCEFILESYSTEM);
            if(SUCCEEDED(spFileOpenDialog->Show(hOwner))) {
                CComPtr<IShellItem> spResult;
                if(SUCCEEDED(spFileOpenDialog->GetResult(&spResult))) {
                    wchar_t* name;
                    if(SUCCEEDED(spResult->GetDisplayName(SIGDN_FILESYSPATH, &name))) {
                        ret = name;
                        CoTaskMemFree(name);
                    }
                }
            }
        }
    }
    return std::move(ret);
}

struct IMessageFilter
{
    virtual bool FilterMessage(MSG* pMsg) = 0;
};

IMessageFilter* pActiveFilter;
size_t nWindowCount = 0;

class PreviewWindow
    : public CWindowImpl<PreviewWindow>
    , public IMessageFilter
{
    BEGIN_MSG_MAP(PreviewWindow)
        MESSAGE_HANDLER(WM_CREATE, OnCreate)
        MESSAGE_HANDLER(WM_DESTROY, OnDestroy)
        MESSAGE_HANDLER(WM_SIZE, OnSize)
        MESSAGE_HANDLER(WM_ACTIVATE, OnActivate)
        MESSAGE_HANDLER(WM_ERASEBKGND, OnEraseBackground)
    END_MSG_MAP()

public:
    PreviewWindow(const std::wstring& content)
        : _content(content)
    { }

    void Create(HWND hOwner)
    {
        int screen_width = ::GetSystemMetrics(SM_CXSCREEN);
        int screen_height = ::GetSystemMetrics(SM_CYSCREEN);
        int my_width = 700;
        int my_height = 600;
        int my_left = (screen_width - my_width) / 2;
        int my_top = (screen_height - my_height) / 2;

        RECT rc {my_left, my_top, my_left + my_width, my_top + my_height};
        CWindowImpl<PreviewWindow>::Create(hOwner, &rc, L"Preview", WS_OVERLAPPEDWINDOW);
    }

protected:
    bool FilterMessage(MSG* pMsg) override
    {
        if(pMsg->message == WM_KEYDOWN && pMsg->wParam == VK_F11) {
            ToggleFullscreen();
        }

        return _pwbc->FilterMessage(pMsg);
    }

    LRESULT OnCreate(UINT uMsg, WPARAM wParam, LPARAM lParam, BOOL& bHandled)
    {
        _pwbc = CreateBroserInstance();
        _pwbc->Create(m_hWnd);
        _pwbc->AddCallable(L"content", BIND(PreviewWindow, OnContent));
        auto index = app_path() + L"\\preview.html";
        _pwbc->Navigate(index.c_str());
        nWindowCount++;
        return 0;
    }

    LRESULT OnDestroy(UINT uMsg, WPARAM wParam, LPARAM lParam, BOOL& bHandled)
    {
        _pwbc->Destroy();
        if(--nWindowCount == 0) {
            ::PostQuitMessage(0);
        }
        return 0;
    }

    LRESULT OnSize(UINT uMsg, WPARAM wParam, LPARAM lParam, BOOL& bHandled)
    {
        RECT rc {0, 0, GET_X_LPARAM(lParam), GET_Y_LPARAM(lParam)};
        _pwbc->SetPos(rc);
        return 0;
    }

    LRESULT OnActivate(UINT uMsg, WPARAM wParam, LPARAM lParam, BOOL& bHandled)
    {
        bool activate = LOWORD(wParam) != WA_INACTIVE;
        if(activate) {
            _pwbc->SetFocus();
            pActiveFilter = this;
        }
        else {
            if(pActiveFilter == this) {
                pActiveFilter = nullptr;
            }
        }

        bHandled = FALSE;
        return 0;
    }

    LRESULT OnEraseBackground(UINT uMsg, WPARAM wParam, LPARAM lParam, BOOL& bHandled)
    {
        return TRUE;
    }

protected:
    taoblog::ComRet OnContent(taoblog::DispParamsVisitor args, VARIANT* result)
    {
        result->vt = VT_BSTR;
        result->bstrVal = ::SysAllocString(_content.c_str());
        return S_OK;
    }

    void ToggleFullscreen()
    {
        LONG_PTR dwStyle = GetWindowLongPtr(GWL_STYLE);
        if(dwStyle & WS_POPUP) {
            dwStyle &= ~WS_POPUP;
            dwStyle |= WS_OVERLAPPEDWINDOW;
            SetWindowLongPtr(GWL_STYLE, dwStyle);
            const auto& r = _rc_restore;
            SetWindowPos(nullptr, r.left, r.top, r.right - r.left, r.bottom - r.top, SWP_FRAMECHANGED);
        }
        else {
            GetWindowRect(&_rc_restore);
            dwStyle &= ~WS_OVERLAPPEDWINDOW;
            dwStyle |= WS_POPUP;
            SetWindowLongPtr(GWL_STYLE, dwStyle);
            SetWindowPos(HWND_TOP, 0, 0, GetSystemMetrics(SM_CXSCREEN), GetSystemMetrics(SM_CYSCREEN), SWP_FRAMECHANGED);
        }
    }

public:

protected:
    IWebBrowserContainer* _pwbc;
    std::wstring _content;
    RECT _rc_restore;
};

class EditorWindow
    : public CWindowImpl<EditorWindow>
    , public IMessageFilter
{
    BEGIN_MSG_MAP(EditorWindow)
        MESSAGE_HANDLER(WM_CREATE, OnCreate)
        MESSAGE_HANDLER(WM_DESTROY, OnDestroy)
        MESSAGE_HANDLER(WM_SIZE, OnSize)
        MESSAGE_HANDLER(WM_ACTIVATE, OnActivate)
        MESSAGE_HANDLER(WM_ERASEBKGND, OnEraseBackground)
    END_MSG_MAP()

public:
    void Create(HWND hOwner)
    {
        int screen_width = ::GetSystemMetrics(SM_CXSCREEN);
        int screen_height = ::GetSystemMetrics(SM_CYSCREEN);
        int my_width = 700;
        int my_height = 600;
        int my_left = (screen_width - my_width) / 2;
        int my_top = (screen_height - my_height) / 2;

        RECT rc {my_left, my_top, my_left + my_width, my_top + my_height};
        CWindowImpl<EditorWindow>::Create(hOwner, &rc, L"Compose", WS_OVERLAPPEDWINDOW);
    }

protected:
    bool FilterMessage(MSG* pMsg) override
    {
        if(pMsg->message == WM_KEYDOWN && pMsg->wParam == VK_F11) {
            ToggleFullscreen();
        }

        return _pwbc->FilterMessage(pMsg);
    }

    LRESULT OnCreate(UINT uMsg, WPARAM wParam, LPARAM lParam, BOOL& bHandled)
    {
        _pwbc = CreateBroserInstance();
        _pwbc->Create(m_hWnd);
        _pwbc->AddCallable(L"open", BIND(EditorWindow, OnOpen));
        _pwbc->AddCallable(L"preview", BIND(EditorWindow, OnPreview));
        _pwbc->AddCallable(L"save", BIND(EditorWindow, OnSave));
        _pwbc->AddCallable(L"export", BIND(EditorWindow, OnExport));
        auto index = app_path() + L"\\compose.html";
        _pwbc->Navigate(index.c_str());
        nWindowCount++;
        return 0;
    }

    LRESULT OnDestroy(UINT uMsg, WPARAM wParam, LPARAM lParam, BOOL& bHandled)
    {
        _pwbc->Destroy();
        if(--nWindowCount == 0) {
            ::PostQuitMessage(0);
        }
        return 0;
    }

    LRESULT OnSize(UINT uMsg, WPARAM wParam, LPARAM lParam, BOOL& bHandled)
    {
        RECT rc {0, 0, GET_X_LPARAM(lParam), GET_Y_LPARAM(lParam)};
        _pwbc->SetPos(rc);
        return 0;
    }

    LRESULT OnActivate(UINT uMsg, WPARAM wParam, LPARAM lParam, BOOL& bHandled)
    {
        bool activate = LOWORD(wParam) != WA_INACTIVE;
        if(activate) {
            _pwbc->SetFocus();
            pActiveFilter = this;
        }
        else {
            if(pActiveFilter == this) {
                pActiveFilter = nullptr;
            }
        }

        bHandled = FALSE;
        return 0;
    }

    LRESULT OnEraseBackground(UINT uMsg, WPARAM wParam, LPARAM lParam, BOOL& bHandled)
    {
        return TRUE;
    }

protected:
    taoblog::ComRet OnOpen(taoblog::DispParamsVisitor args, VARIANT* result)
    {
        auto path = file_open_dialog(m_hWnd);
        if(path.empty()) return E_ABORT;

        std::ifstream file(path, std::ios::binary);
        if(!file.is_open()) return E_FAIL;

        _curr_dir = path.substr(0, path.find_last_of(L"/\\"));

        file.seekg(0, std::ios::end);
        auto size = (size_t)file.tellg();
        file.seekg(0, std::ios::beg);
        auto buf = std::make_unique<char[]>(size+1);
        file.read(buf.get(), size);
        buf.get()[size] = L'\0';

        auto wbuf = U8U16Cvt().from_bytes(buf.get());

        result->vt = VT_BSTR;
        result->bstrVal = ::SysAllocString(wbuf.c_str());
        return S_OK;
    }

    taoblog::ComRet OnPreview(taoblog::DispParamsVisitor args, VARIANT* result)
    {
        auto prewnd = new PreviewWindow(args[0].bstrVal);
        prewnd->Create(m_hWnd);
        prewnd->ShowWindow(SW_SHOWNORMAL);
        return S_OK;
    }

    taoblog::ComRet OnSave(taoblog::DispParamsVisitor args, VARIANT* result)
    {
        if(_curr_dir.empty()) {
            MessageBox(L"没有选择文件。");
            return E_ABORT;
        }

        auto wstr = std::wstring(args[0].bstrVal);
        auto str = U8U16Cvt().to_bytes(wstr);
        auto path = _curr_dir + L"\\README.md";
        std::ofstream file(path, std::ios::binary | std::ios::trunc);
        file << str;
        file.close();
        return S_OK;
    }

    taoblog::ComRet OnExport(taoblog::DispParamsVisitor args, VARIANT* result)
    {
        if(_curr_dir.empty()) {
            MessageBox(L"没有选择文件。");
            return E_ABORT;
        }

        auto wstr = std::wstring(args[0].bstrVal);
        auto str = U8U16Cvt().to_bytes(wstr);
        auto path = _curr_dir + L"\\content.html";
        std::ofstream file(path, std::ios::binary | std::ios::trunc);
        file << str;
        file.close();
        return S_OK;
    }

    void ToggleFullscreen()
    {
        LONG_PTR dwStyle = GetWindowLongPtr(GWL_STYLE);
        if(dwStyle & WS_POPUP) {
            dwStyle &= ~WS_POPUP;
            dwStyle |= WS_OVERLAPPEDWINDOW;
            SetWindowLongPtr(GWL_STYLE, dwStyle);
            const auto& r = _rc_restore;
            SetWindowPos(nullptr, r.left, r.top, r.right - r.left, r.bottom - r.top, SWP_FRAMECHANGED);
        }
        else {
            GetWindowRect(&_rc_restore);
            dwStyle &= ~WS_OVERLAPPEDWINDOW;
            dwStyle |= WS_POPUP;
            SetWindowLongPtr(GWL_STYLE, dwStyle);
            SetWindowPos(HWND_TOP, 0, 0, GetSystemMetrics(SM_CXSCREEN), GetSystemMetrics(SM_CYSCREEN), SWP_FRAMECHANGED);
        }
    }
public:

protected:
    IWebBrowserContainer* _pwbc;
    RECT _rc_restore;
    std::wstring _curr_dir;
};

}

int __stdcall wWinMain(HINSTANCE hInstance, HINSTANCE, LPWSTR lpCmdLine, int nShowCmd)
{
    CoInitialize(nullptr);

    taoblog::WebBrowserVersionSetter _verset;

    auto edtwnd = new taoblog::EditorWindow;
    edtwnd->Create(nullptr);
    edtwnd->ShowWindow(nShowCmd);

    MSG msg;
    while(::GetMessage(&msg, nullptr, 0, 0))
    {
        if(!taoblog::pActiveFilter || !taoblog::pActiveFilter->FilterMessage(&msg)) {
            ::TranslateMessage(&msg);
            ::DispatchMessage(&msg);
        }
    }

    return (int)msg.wParam;
}
