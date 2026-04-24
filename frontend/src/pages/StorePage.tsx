// StorePage.tsx — Halaman utama game store selepas user login.
//
// Tanggungjawab halaman ni:
//   1. Fetch produk dari Xsolla Store API (guna access token user)
//   2. Fetch senarai item yang user dah beli dari backend kita
//   3. Urus cart (tambah, kurang, buang item)
//   4. Handle checkout → hantar ke backend → redirect ke Xsolla Pay Station
//   5. Detect balik dari Pay Station (URL ?payment=success) dan refresh inventory
//
// Files backend yang terlibat:
//   - GET /purchases       → internal/handler/purchase.go → GetPurchases()
//   - POST /create-payment → internal/handler/payment.go  → CreatePayment()

import { useEffect, useMemo, useState } from 'react'
import { useNavigate, useLocation } from 'react-router-dom'
import '../styles/store.css'

// ── Type definitions ──────────────────────────────────────────────────────────

type Item = {
  id: string
  sku: string
  name: string
  description: string
  price: number
  currency: string
  image_url: string
}

type CartItem = Item & {
  quantity: number
}

type InventoryItem = Item & {
  quantity: number
}

export default function StorePage() {
  const navigate = useNavigate()
  const location = useLocation()

  // ── State variables ─────────────────────────────────────────────────────────
  const [items, setItems] = useState<Item[]>([])           // Produk dari Xsolla catalog
  const [cart, setCart] = useState<CartItem[]>([])         // Item dalam cart (state tempatan je)
  const [userEmail, setUserEmail] = useState('player@gardenwars.com') // Email untuk header
  const [isCartOpen, setIsCartOpen] = useState(false)      // Toggle drawer cart
  const [paymentSuccess, setPaymentSuccess] = useState(false) // Toast selepas bayar
  const [purchasedSKUs, setPurchasedSKUs] = useState<string[]>([]) // SKU yang dah dibeli
  const [activeTab, setActiveTab] = useState<'store' | 'inventory'>('store') // Tab aktif

  // ── Konstanta ───────────────────────────────────────────────────────────────
  const LIMITED_SKU = 'dancing-zombie'                              // SKU item yang hanya boleh dibeli SEKALI
  const BACKEND_URL = import.meta.env.VITE_API_BASE_URL as string  // URL backend (dari .env)
  const XSOLLA_PROJECT_ID = '304864'                                // ID project Xsolla store

  // ── Detect balik dari Pay Station ───────────────────────────────────────────
  // Xsolla redirect ke /store?payment=success selepas bayar berjaya
  // Kita check URL parameter ni dan refresh inventory user
  useEffect(() => {
    const params = new URLSearchParams(location.search)
    if (params.get('payment') === 'success') {
      setPaymentSuccess(true)        // Tunjuk toast "Payment successful"
      setCart([])                    // Clear cart
      window.history.replaceState({}, '', '/store') // Buang ?payment=success dari URL
      fetchPurchases()               // Refresh senarai item yang dah dibeli
    }
  }, [location.search])

  // ── fetchPurchases — ambil senarai SKU yang user dah beli dari backend ───────
  // Backend: GET /purchases → internal/handler/purchase.go → GetPurchases()
  // Data ni digunakan untuk: (1) tunjuk inventory, (2) disable butang beli limited item
  const fetchPurchases = async () => {
    const token = sessionStorage.getItem('xsolla_access_token')
    if (!token) {
      console.warn('No access token found, skip fetchPurchases')
      return
    }

    try {
      const res = await fetch(`${BACKEND_URL}/purchases`, {
        headers: { Authorization: `Bearer ${token}` },
      })

      if (!res.ok) {
        console.error('Failed purchases response:', res.status)
        return
      }

      const data = await res.json()
      setPurchasedSKUs(data.purchased_skus || [])
    } catch (error) {
      console.error('Failed to fetch purchases:', error)
    }
  }

  // ── On page load — fetch email, purchases, dan produk ──────────────────────
  useEffect(() => {
    // Ambil email dari sessionStorage (disimpan semasa login dalam AuthCallbackPage)
    const savedEmail = sessionStorage.getItem('user_email')
    if (savedEmail) setUserEmail(savedEmail)

    // Fetch senarai item yang user dah beli
    fetchPurchases()

    // Fetch catalog produk dari Xsolla Store API menggunakan access token user
    // Kenapa guna token user? Sebab Xsolla tunjuk harga dan mata wang ikut lokasi user
    const fetchItems = async () => {
      try {
        const token = sessionStorage.getItem('xsolla_access_token')
        if (!token) {
          throw new Error('Xsolla access token not found')
        }

        // Xsolla Store API — return semua virtual items untuk project kita
        const res = await fetch(
          `https://store.xsolla.com/api/v2/project/${XSOLLA_PROJECT_ID}/items`,
          {
            method: 'GET',
            headers: {
              Authorization: `Bearer ${token}`,
              'Content-Type': 'application/json',
            },
          }
        )

        if (!res.ok) {
          const text = await res.text()
          console.error('Catalog error:', text)
          throw new Error('Failed to fetch publisher catalog')
        }

        const data = await res.json()

        const mappedItems: Item[] = (data.items || []).map((item: any) => ({
          id: item.sku,
          sku: item.sku,
          name: item.name,
          description: item.description || '',
          price: Number(item.price?.amount || 0),
          currency: item.price?.currency || 'USD',
          image_url: item.image_url || '',
        }))

        setItems(mappedItems)
      } catch (error) {
        console.error('Failed to fetch items from publisher:', error)
        setItems([])
      }
    }

    fetchItems()
  }, [])

  // ── Computed values (useMemo — recalculate only when dependencies change) ────

  // cartCount — jumlah total unit dalam cart (untuk badge pada ikon cart)
  const cartCount = useMemo(() => {
    return cart.reduce((total, item) => total + item.quantity, 0)
  }, [cart])

  // cartTotal — jumlah harga semua item dalam cart
  const cartTotal = useMemo(() => {
    return cart.reduce((total, item) => total + item.price * item.quantity, 0)
  }, [cart])

  // inventoryItems — transform senarai SKU yang dibeli jadi array item dengan quantity
  // Group purchasedSKUs by sku using reduce() — one card per item with quantity
  const inventoryItems = useMemo((): InventoryItem[] => {
    const grouped = purchasedSKUs.reduce(
      (acc, sku) => {
        if (acc[sku]) {
          acc[sku].quantity += 1
        } else {
          const item = items.find((i) => i.sku === sku)
          if (item) acc[sku] = { ...item, quantity: 1 }
        }
        return acc
      },
      {} as Record<string, InventoryItem>
    )
    return Object.values(grouped)
  }, [purchasedSKUs, items])

  // ── Cart & item logic ────────────────────────────────────────────────────────

  // isItemDisabled — semak sama ada butang "Add to Cart" perlu di-disable
  // Hanya untuk dancing-zombie (LIMITED_SKU): disable kalau dah beli atau dah dalam cart
  const isItemDisabled = (sku: string): boolean => {
    if (sku !== LIMITED_SKU) return false

    const alreadyPurchased = purchasedSKUs.includes(sku)
    const alreadyInCart = cart.some((c) => c.sku === sku)

    return alreadyPurchased || alreadyInCart
  }

  // addToCart — tambah item ke cart
  // Untuk limited item: hanya boleh ada satu unit, tak boleh tambah lagi
  const addToCart = (item: Item) => {
    if (isItemDisabled(item.sku)) return

    setCart((prevCart) => {
      const existingItem = prevCart.find((cartItem) => cartItem.id === item.id)

      if (existingItem) {
        if (item.sku === LIMITED_SKU) return prevCart

        return prevCart.map((cartItem) =>
          cartItem.id === item.id
            ? { ...cartItem, quantity: cartItem.quantity + 1 }
            : cartItem
        )
      }

      return [...prevCart, { ...item, quantity: 1 }]
    })

    setIsCartOpen(true)
  }

  // increaseQuantity — naikkan quantity item dalam cart (skip kalau limited item)
  const increaseQuantity = (id: string) => {
    setCart((prevCart) =>
      prevCart.map((item) => {
        if (item.id === id && item.sku === LIMITED_SKU) return item
        return item.id === id ? { ...item, quantity: item.quantity + 1 } : item
      })
    )
  }

  // decreaseQuantity — kurangkan quantity; kalau jadi 0, item dibuang dari cart
  const decreaseQuantity = (id: string) => {
    setCart((prevCart) =>
      prevCart
        .map((item) =>
          item.id === id ? { ...item, quantity: item.quantity - 1 } : item
        )
        .filter((item) => item.quantity > 0)
    )
  }

  // removeItem — buang item sepenuhnya dari cart
  const removeItem = (id: string) => {
    setCart((prevCart) => prevCart.filter((item) => item.id !== id))
  }

  // ── Auth actions ─────────────────────────────────────────────────────────────

  // handleLogout — buang semua data dari sessionStorage dan redirect ke login
  const handleLogout = () => {
    sessionStorage.removeItem('xsolla_state')
    sessionStorage.removeItem('user_email')
    sessionStorage.removeItem('user_name')
    sessionStorage.removeItem('xsolla_code')
    sessionStorage.removeItem('xsolla_access_token')
    sessionStorage.removeItem('xsolla_refresh_token')
    sessionStorage.removeItem('xsolla_provider')
    sessionStorage.removeItem('xsolla_user_id')
    sessionStorage.removeItem('xsolla_picture')
    navigate('/')
  }

  // ── handleCheckout — proses pembayaran ──────────────────────────────────────
  // Flow:
  //   1. Validate cart dan token
  //   2. Hantar ke backend POST /create-payment
  //   3. Backend buat cart dalam Xsolla dan return payment token
  //   4. Redirect ke Xsolla Pay Station dengan token tu
  //   5. Selepas bayar, Xsolla redirect ke /store?payment=success
  //
  // Backend file: internal/handler/payment.go → CreatePayment()
  const handleCheckout = async () => {
    try {
      if (cart.length === 0) return

      // Ambil token dan user ID yang disimpan semasa login
      const userToken = sessionStorage.getItem('xsolla_access_token')
      const userID = sessionStorage.getItem('xsolla_user_id')

      if (!userToken || !userID) {
        alert('Please login first with Xsolla before checkout.')
        return
      }

      // Ambil currency dari item pertama dalam cart
      // Xsolla return semua item dalam currency yang sama untuk satu user (ikut lokasi)
      const currency = cart[0]?.currency || 'USD'

      // Bina payload: senarai item dalam cart + token user untuk Xsolla Store API
      const payload = {
        items: cart.map((item) => ({
          sku: item.sku,
          quantity: item.quantity,
        })),
        currency,
        user_token: userToken,
        user_id: userID,
      }

      const res = await fetch(`${BACKEND_URL}/create-payment`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${userToken}`,
        },
        body: JSON.stringify(payload),
      })

      if (!res.ok) {
        const text = await res.text()
        alert(`Checkout failed: ${text}`)
        return
      }

      const data = await res.json()

      if (!data.token) {
        alert('Xsolla token not found')
        return
      }

      // Redirect ke Xsolla Pay Station sandbox dengan token yang diterima dari backend
      window.location.href = `https://sandbox-secure.xsolla.com/paystation4/?token=${data.token}`
    } catch (error) {
      alert('Checkout failed')
    }
  }

  return (
    <div className="store-page">
      {paymentSuccess && (
        <div className="payment-toast" role="alert">
          <div className="payment-toast__body">
            <p className="payment-toast__title">Payment success</p>
            <p className="payment-toast__msg">
              Payment successful. Your cart is cleared.
              <br />
              You can stay here for a while, or close this message.
            </p>
          </div>
          <button
            className="payment-toast__close"
            type="button"
            onClick={() => setPaymentSuccess(false)}
          >
            Close
          </button>
        </div>
      )}

      <header className="store-header">
        <div className="store-header__user">
          <div className="store-header__avatar">
            {userEmail.charAt(0).toUpperCase()}
          </div>
          <div className="store-header__user-info">
            <span className="store-header__label">Logged in as</span>
            <span className="store-header__email">{userEmail}</span>
          </div>
        </div>

        <div className="store-header__actions">
          <button
            className="cart-button"
            type="button"
            onClick={() => setIsCartOpen(true)}
          >
            🛒
            <span className="cart-button__count">{cartCount}</span>
          </button>

          <button className="logout-button" type="button" onClick={handleLogout}>
            Logout
          </button>
        </div>
      </header>

      <main className="store-main">
        <section className="store-hero">
          <h1 className="store-hero__title">Garden Wars Item Shop</h1>
          <p className="store-hero__subtitle">
            Discover powerful Plants vs Zombies items for your defense strategy.
          </p>
        </section>

        {/* ── TABS ── */}
        <div className="tabs-wrapper">
          <div className="tabs">
            <button
              className={`tab-btn${activeTab === 'store' ? ' tab-btn--active' : ''}`}
              type="button"
              onClick={() => setActiveTab('store')}
            >
              Store
            </button>
            <button
              className={`tab-btn${activeTab === 'inventory' ? ' tab-btn--active' : ''}`}
              type="button"
              onClick={() => setActiveTab('inventory')}
            >
              My Inventory
            </button>
          </div>
        </div>

        {/* ── STORE TAB ── */}
        {activeTab === 'store' && (
          <section className="store-grid">
            {items.map((item) => {
              const disabled = isItemDisabled(item.sku)
              const alreadyPurchased = purchasedSKUs.includes(item.sku)
              const isLimited = item.sku === LIMITED_SKU

              return (
                <article className="item-card" key={item.id}>
                  <div className="item-card__image-wrap">
                    <img
                      src={item.image_url}
                      alt={item.name}
                      className="item-card__image"
                    />
                  </div>

                  <div className="item-card__content">
                    <h3 className="item-card__name">{item.name}</h3>
                    <p className="item-card__description">{item.description}</p>
                    <p className="item-card__price">
                      {item.currency} {Number(item.price).toFixed(2)}
                    </p>

                    <button
                      className="item-card__button"
                      type="button"
                      onClick={() => addToCart(item)}
                      disabled={disabled}
                      style={disabled ? { opacity: 0.45, cursor: 'not-allowed' } : undefined}
                    >
                      {isLimited && alreadyPurchased
                        ? 'Already Purchased'
                        : isLimited && disabled
                        ? 'Already in Cart'
                        : 'Add to Cart'}
                    </button>
                  </div>
                </article>
              )
            })}
          </section>
        )}

        {/* ── INVENTORY TAB ── */}
        {activeTab === 'inventory' && (
          <section className="store-grid">
            {inventoryItems.length === 0 ? (
              <div className="inventory-empty">
                <h3 className="inventory-empty__title">Your inventory is empty</h3>
                <p className="inventory-empty__msg">
                  Head to the Store to purchase some items.
                </p>
              </div>
            ) : (
              inventoryItems.map((item) => (
                <article className="item-card" key={item.sku}>
                  {item.quantity > 1 && (
                    <div className="quantity-badge">x{item.quantity}</div>
                  )}
                  <div className="item-card__image-wrap">
                    <img
                      src={item.image_url}
                      alt={item.name}
                      className="item-card__image"
                    />
                  </div>
                  <div className="item-card__content">
                    <h3 className="item-card__name">{item.name}</h3>
                    <p className="item-card__description">{item.description}</p>
                    <p className="item-card__price">
                      {item.currency} {Number(item.price).toFixed(2)}
                    </p>
                  </div>
                </article>
              ))
            )}
          </section>
        )}
      </main>

      {isCartOpen && (
        <>
          <div className="cart-overlay" onClick={() => setIsCartOpen(false)} />

          <aside className="cart-drawer">
            <div className="cart-drawer__header">
              <h2 className="cart-drawer__title">Your Cart</h2>
              <button
                className="cart-drawer__close"
                type="button"
                onClick={() => setIsCartOpen(false)}
              >
                ✕
              </button>
            </div>

            <div className="cart-drawer__body">
              {cart.length === 0 ? (
                <p className="cart-empty">Your cart is empty.</p>
              ) : (
                cart.map((item) => (
                  <div className="cart-item" key={item.id}>
                    <img
                      src={item.image_url}
                      alt={item.name}
                      className="cart-item__image"
                    />
                    <div className="cart-item__info">
                      <h4 className="cart-item__name">{item.name}</h4>
                      <p className="cart-item__price">
                        {item.currency} {Number(item.price).toFixed(2)}
                      </p>

                      <div className="cart-item__actions">
                        <button type="button" onClick={() => decreaseQuantity(item.id)}>
                          -
                        </button>

                        <span>{item.quantity}</span>

                        <button
                          type="button"
                          onClick={() => increaseQuantity(item.id)}
                          disabled={item.sku === LIMITED_SKU}
                          style={
                            item.sku === LIMITED_SKU
                              ? { opacity: 0.3, cursor: 'not-allowed' }
                              : undefined
                          }
                        >
                          +
                        </button>
                      </div>
                    </div>

                    <button
                      className="cart-item__remove"
                      type="button"
                      onClick={() => removeItem(item.id)}
                    >
                      Remove
                    </button>
                  </div>
                ))
              )}
            </div>

            <div className="cart-drawer__footer">
              <div className="cart-total">
                <span>Total</span>
                <strong>{cart[0]?.currency || 'USD'} {cartTotal.toFixed(2)}</strong>
              </div>

              <button
                className="checkout-button"
                type="button"
                onClick={handleCheckout}
                disabled={cart.length === 0}
              >
                Checkout
              </button>
            </div>
          </aside>
        </>
      )}
    </div>
  )
}
