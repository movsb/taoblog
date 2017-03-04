#pragma once

#include <windows.h>

namespace taoblog {

class ComRet
{
public:
    ComRet() : _hr(E_FAIL) {}
    ComRet(HRESULT hr) : _hr(hr) {}

    operator bool() const { return SUCCEEDED(_hr); }
    operator HRESULT() const { return _hr; }

protected:
    HRESULT _hr;
};

class DispParamsVisitor
{
public:
    DispParamsVisitor(UINT argc, VARIANT* argv)
        : _argc(argc)
        , _argv(argv)
    {

    }

    VARIANT& operator[](UINT i)
    {
        return _argv[_argc - i - 1];
    }

    UINT size() const
    {
        return _argc;
    }

protected:
    UINT        _argc;
    VARIANT*    _argv;
};

template<class T>
class ComPtrBase
{
public:
    ComPtrBase()
        : _p(nullptr)
    {
    }

    ComPtrBase(T* p)
    {
        _p = p;
        if(_p) {
            _p->AddRef();
        }
    }

    ~ComPtrBase()
    {
        if(_p) {
            _p->Release();
            _p = nullptr;
        }
    }

public:
    operator T*() const
    {
        return _p;
    }

    T& operator*() const
    {
        return *_p;
    }

    T** operator&()
    {
        return &_p;
    }

    T* operator->()
    {
        return _p;
    }

    bool operator!() const
    {
        return !_p;
    }

    bool operator==(T* p) const
    {
        return _p == p;
    }

    bool operator!=(T* p) const
    {
        return !operator==(p);
    }

public:
    void Release()
    {
        if(_p) {
            _p->Release();
            _p = nullptr;
        }
    }

    T* Detach()
    {
        T* t = _p;
        _p = nullptr;
        return t;
    }

    HRESULT CopyTo(T** ppT)
    {
        *ppT = _p;
        _p && _p->AddRef();
        return S_OK;
    }

    template<class Q>
    HRESULT QueryInterface(Q** pp) const
    {
        return _p->QeuryInterface(__uuidof(Q), (void**)pp);
    }

    bool IsEqualObject(IUnknown* pOther)
    {
        ComPtr<IUnknown> p1;
        ComPtr<IUnknown> p2;

        _p->QueryInterface(__uuidof(IUnknown), (void**)&p1);
        pOther->QueryInterface(__uuidof(IUnknown), (void**)&p2);

        return p1 == p2;
    }

protected:
    T* _p;
};

template<class T>
class ComPtr : public ComPtrBase<T>
{
public:
    T* operator=(T* p)
    {
        if(_p != p) {
            if(_p) _p->Release();
            _p = p;
            if(_p) _p->AddRef();
        }
        return _p;
    }

    T* operator=(const ComPtr<T>& p)
    {
        return operator=(p._p);
    }
};

template<>
class ComPtr<IDispatch>
    : public ComPtrBase<IDispatch>
{
public:
    ComPtr()
        : ComPtrBase<IDispatch>(nullptr)
    {

    }

    ComPtr(IDispatch* disp)
        : ComPtrBase<IDispatch>(disp)
    {

    }

    HRESULT GetIDOfName(LPCOLESTR name, DISPID* id)
    {
        return _p->GetIDsOfNames(IID_NULL, const_cast<LPOLESTR*>(&name), 1, LOCALE_USER_DEFAULT, id);
    }

    HRESULT PutProperty(DISPID id, VARIANT* var)
    {
        return PutProperty(_p, id, var);
    }

    HRESULT PutProperty(LPCOLESTR name, VARIANT* var)
    {
        DISPID id;
        HRESULT hr = GetIDOfName(name, &id);
        if(SUCCEEDED(hr))
            hr = PutProperty(id, var);
        return hr;
    }

    HRESULT GetProperty(DISPID id, VARIANT* var)
    {
        return GetProperty(_p, id, var);
    }

    HRESULT GetProperty(LPCOLESTR name, VARIANT* var)
    {
        DISPID id;
        HRESULT hr;

        hr = GetIDOfName(name, &id);
        if(SUCCEEDED(hr)) {
            hr = GetProperty(id, var);
        }

        return hr;
    }

public:
    HRESULT Invoke(DISPID id, VARIANT* ret)
    {
        DISPPARAMS args = {nullptr, nullptr, 0, 0};
        return _p->Invoke(id, IID_NULL, LOCALE_USER_DEFAULT, DISPATCH_METHOD, &args, ret, nullptr, nullptr);
    }

    HRESULT Invoke(LPCOLESTR name, VARIANT* ret)
    {
        HRESULT hr;
        DISPID id;
        
        hr = GetIDOfName(name, &id);
        if(SUCCEEDED(hr))
            hr = Invoke(id, ret);

        return hr;
    }

    HRESULT Invoke(DISPID id, VARIANT* v1, VARIANT* ret)
    {
        DISPPARAMS args = {v1, nullptr, 1, 0};
        return _p->Invoke(id, IID_NULL, LOCALE_USER_DEFAULT, DISPATCH_METHOD, &args, ret, nullptr, nullptr);
    }

    HRESULT Invoke(LPCOLESTR name, VARIANT* v1, VARIANT* ret)
    {
        DISPID id;
        ComRet hr;

        hr = GetIDOfName(name, &id);
        if(hr)
            hr = Invoke(id, v1, ret);

        return hr;
    }

    HRESULT Invoke(DISPID id, UINT argc, VARIANT* argv, VARIANT* ret)
    {
        DISPPARAMS args = {argv, nullptr, argc, 0};
        return _p->Invoke(id, IID_NULL, LOCALE_USER_DEFAULT, DISPATCH_METHOD, &args, ret, nullptr, nullptr);
    }

    HRESULT Invoke(LPCOLESTR name, UINT argc, VARIANT* argv, VARIANT* ret)
    {
        DISPID id;
        ComRet hr;

        hr = GetIDOfName(name, &id);
        if(hr) {
            hr = Invoke(id, argc, argv, ret);
        }

        return hr;
    }

protected:
    static HRESULT PutProperty(IDispatch* pDisp, DISPID id, VARIANT* var)
    {
        DISPPARAMS args = {nullptr, nullptr, 1, 1};
        args.rgvarg = var;
        DISPID idPut = DISPID_PROPERTYPUT;
        args.rgdispidNamedArgs = &idPut;

        if(var->vt == VT_UNKNOWN || var->vt == VT_DISPATCH || var->vt & VT_ARRAY || var->vt & VT_BYREF) {
            HRESULT hr = pDisp->Invoke(id, IID_NULL, LOCALE_USER_DEFAULT, DISPATCH_PROPERTYPUTREF, &args, nullptr, nullptr, nullptr);
            if(SUCCEEDED(hr))
                return hr;
        }

        return  pDisp->Invoke(id, IID_NULL, LOCALE_USER_DEFAULT, DISPATCH_PROPERTYPUT, &args, nullptr, nullptr, nullptr);
    }

    static HRESULT GetProperty(IDispatch* pDisp, DISPID id, VARIANT* var)
    {
        DISPPARAMS args = {nullptr, nullptr, 0, 0};
        return pDisp->Invoke(id, IID_NULL, LOCALE_USER_DEFAULT, DISPATCH_PROPERTYGET, &args, var, nullptr, nullptr);
    }
};

template<class T, const IID* piid = &__uuidof(T)>
class ComQIPtr : public ComPtrBase<T>
{
public:
    ComQIPtr()
    { }

    ComQIPtr(T* p)
        : ComPtrBase<T>(p)
    { }

    ComQIPtr(IUnknown* p)
    {
        if(FAILED(p->QueryInterface(*piid, (void**)&_p)))
            _p = nullptr;
    }
};

class ComVariant : public tagVARIANT
{
public:
    ComVariant()
    {
        memset(this, 0, sizeof(tagVARIANT));
        ::VariantInit(this);
    }

    ComVariant(LPCWSTR s)
    {
        vt = VT_EMPTY;
        *this = s;
    }

    ComVariant(bool b)
    {
        vt = VT_BOOL;
        boolVal = b ? VARIANT_TRUE : VARIANT_FALSE;
    }

    ComVariant(int i)
    {
        vt = VT_I4;
        intVal = i;
    }

    ComVariant(IDispatch* p)
    {
        vt = VT_DISPATCH;
        pdispVal = p;

        pdispVal && pdispVal->AddRef();
    }

    ComVariant(IUnknown* p)
    {
        vt = VT_UNKNOWN;
        punkVal = p;

        punkVal && punkVal->AddRef();
    }

public:
    ComVariant& operator=(LPCOLESTR s)
    {
        if(vt != VT_BSTR || bstrVal != s) {
            Clear();

            vt = VT_BSTR;
            bstrVal = ::SysAllocString(s);

            if(bstrVal == nullptr && s != nullptr) {
                vt = VT_ERROR;
            }
        }

        return *this;
    }

    ComVariant& operator=(bool b)
    {
        if(vt != VT_BOOL) {
            Clear();
            vt = VT_BOOL;
        }

        boolVal = b ? VARIANT_TRUE : VARIANT_FALSE;

        return *this;
    }

    ComVariant& operator=(int i)
    {
        if(vt != VT_I4) {
            Clear();
            vt = VT_I4;
        }

        intVal = i;

        return *this;
    }

    ComVariant& operator=(IDispatch* p)
    {
        if(vt != VT_DISPATCH || p != pdispVal) {
            Clear();

            vt = VT_DISPATCH;
            pdispVal = p;

            pdispVal && pdispVal->AddRef();
        }
    }

    ComVariant& operator=(IUnknown* p)
    {
        if(vt != VT_UNKNOWN || punkVal != p) {
            Clear();

            vt = VT_UNKNOWN;
            punkVal = p;

            punkVal && punkVal->AddRef();
        }
    }

public:
    HRESULT Clear()
    {
        return ::VariantClear(this);
    }
};

}
