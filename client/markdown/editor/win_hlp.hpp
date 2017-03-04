#pragma once

namespace taowin
{

template<typename T, typename BOOLER, typename CLOSER>
class AutoCloseT
{
public:
    AutoCloseT(T t) : value(t) {}
    ~AutoCloseT() { BOOLER()(value) && CLOSER()(value); }
    operator bool() const { return BOOLER()(value); }
    operator T() const { return value; }

protected:
    T value;
};

template<typename R, typename T, typename BOOLER, typename CLOSER>
class AutoCloseT2
{
public:
    AutoCloseT2(R r) : rvalue(r) {}
    ~AutoCloseT2() { BOOLER()(rvalue) && CLOSER()(tvalue); }
    operator bool() const { return BOOLER()(rvalue); }
    operator T() const { return tvalue; }
    operator R() const { return rvalue; }
    T* operator &() { return &tvalue; }

protected:
    T tvalue;
    R rvalue;
};

struct HandleBooler { bool operator()(HANDLE h) const { return h != nullptr && h != INVALID_HANDLE_VALUE; } };
struct HandleCloser { bool operator()(HANDLE h) const { return !!::CloseHandle(h); } };
typedef AutoCloseT<HANDLE, HandleBooler, HandleCloser> AutoHandle;

struct RegistryBooler { bool operator()(LRESULT lr) const { return lr == ERROR_SUCCESS; } };
struct RegistryCloser { bool operator()(HKEY hKey) const { return ::RegCloseKey(hKey) == ERROR_SUCCESS; } };
typedef AutoCloseT2<LRESULT, HKEY, RegistryBooler, RegistryCloser> AutoRegistry;

class BoolVal
{
public:
    BoolVal()       : _b(false) {}
    BoolVal(BOOL b) : _b(b) {}
    BoolVal(bool b) : _b(b) {}

    operator BOOL() const { return _b; }
    operator bool() const { return !!_b; }

protected:
    BOOL _b;
};

template<typename R = size_t, typename T = std::wstring>
R string_bytes(const T& t, bool inc_nul = true)
{
    return (R)((t.size() + (inc_nul ? 1 : 0)) * sizeof(t[0]));
}

}
