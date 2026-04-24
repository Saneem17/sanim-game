import { useEffect, useMemo, useState } from 'react'
import { useNavigate, useLocation } from 'react-router-dom'
import '../styles/store.css'
import '../styles/store-overrides.css'

type Item = {
  id: string
  sku: string
  name: string
  description: string
  price: number
  currency: string
  image_url: string
}

type CartItem = Item & { quantity: number }
type InventoryItem = Item & { quantity: number }

export default function StorePage() {
  const navigate = useNavigate()
  const location = useLocation()

  const [items, setItems] = useState<Item[]>([])
  const [cart, setCart] = useState<CartItem[]>([])
  const [userEmail, setUserEmail] = useState('player@gardenwars.com')
  const [isCartOpen, setIsCartOpen] = useState(false)
  const [paymentSuccess, setPaymentSuccess] = useState(false)
  const [purchasedSKUs, setPurchasedSKUs] = useState<string[]>([])
  const [activeTab, setActiveTab] = useState<'store' | 'inventory'>('store')
  const [logoutNotification, setLogoutNotification] = useState(false)

  const LIMITED_SKU = 'dancing-zombie'
  const BACKEND_URL = import.meta.env.VITE_API_BASE_URL as string
  const XSOLLA_PROJECT_ID = '304864'

  useEffect(() => {
    const params = new URLSearchParams(location.search)
    if (params.get('payment') === 'success') {
      setPaymentSuccess(true)
      setCart([])
      window.history.replaceState({}, '', '/store')
      fetchPurchases()
    }
  }, [location.search])

  const fetchPurchases = async () => {
    const token = sessionStorage.getItem('xsolla_access_token')
    if (!token) return
    try {
      const res = await fetch(`${BACKEND_URL}/purchases`, {
        headers: { Authorization: `Bearer ${token}` },
      })
      if (!res.ok) return
      const data = await res.json()
      setPurchasedSKUs(data.purchased_skus || [])
    } catch (error) {
      console.error('Failed to fetch purchases:', error)
    }
  }

  useEffect(() => {
    const savedEmail = sessionStorage.getItem('user_email')
    if (savedEmail) setUserEmail(savedEmail)
    fetchPurchases()

    const fetchItems = async () => {
      try {
        const token = sessionStorage.getItem('xsolla_access_token')
        if (!token) throw new Error('Xsolla access token not found')

        const res = await fetch(
          `https://store.xsolla.com/api/v2/project/${XSOLLA_PROJECT_ID}/items`,
          {
            headers: {
              Authorization: `Bearer ${token}`,
              'Content-Type': 'application/json',
            },
          }
        )
        if (!res.ok) throw new Error('Failed to fetch catalog')
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
        console.error('Failed to fetch items:', error)
        setItems([])
      }
    }

    fetchItems()
  }, [])

  const cartCount = useMemo(() => cart.reduce((t, i) => t + i.quantity, 0), [cart])
  const cartTotal = useMemo(() => cart.reduce((t, i) => t + i.price * i.quantity, 0), [cart])

  const inventoryItems = useMemo((): InventoryItem[] => {
    const grouped = purchasedSKUs.reduce((acc, sku) => {
      if (acc[sku]) {
        acc[sku].quantity += 1
      } else {
        const item = items.find((i) => i.sku === sku)
        if (item) acc[sku] = { ...item, quantity: 1 }
      }
      return acc
    }, {} as Record<string, InventoryItem>)
    return Object.values(grouped)
  }, [purchasedSKUs, items])

  const isItemDisabled = (sku: string): boolean => {
    if (sku !== LIMITED_SKU) return false
    return purchasedSKUs.includes(sku) || cart.some((c) => c.sku === sku)
  }

  const addToCart = (item: Item) => {
    if (isItemDisabled(item.sku)) return
    setCart((prev) => {
      const existing = prev.find((c) => c.id === item.id)
      if (existing) {
        if (item.sku === LIMITED_SKU) return prev
        return prev.map((c) => c.id === item.id ? { ...c, quantity: c.quantity + 1 } : c)
      }
      return [...prev, { ...item, quantity: 1 }]
    })
    setIsCartOpen(true)
  }

  const increaseQuantity = (id: string) => {
    setCart((prev) =>
      prev.map((item) => {
        if (item.id === id && item.sku === LIMITED_SKU) return item
        return item.id === id ? { ...item, quantity: item.quantity + 1 } : item
      })
    )
  }

  const decreaseQuantity = (id: string) => {
    setCart((prev) =>
      prev.map((i) => i.id === id ? { ...i, quantity: i.quantity - 1 } : i)
          .filter((i) => i.quantity > 0)
    )
  }

  const removeItem = (id: string) => {
    setCart((prev) => prev.filter((i) => i.id !== id))
  }

  const handleLogout = () => {
    setLogoutNotification(true)
    setTimeout(() => {
      sessionStorage.clear()
      navigate('/')
    }, 2200)
  }

  const handleCheckout = async () => {
    try {
      if (cart.length === 0) return
      const userToken = sessionStorage.getItem('xsolla_access_token')
      const userID = sessionStorage.getItem('xsolla_user_id')
      if (!userToken || !userID) {
        alert('Please login first with Xsolla before checkout.')
        return
      }
      const currency = cart[0]?.currency || 'USD'
      const payload = {
        items: cart.map((i) => ({ sku: i.sku, quantity: i.quantity })),
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
      if (!data.token) { alert('Xsolla token not found'); return }
      window.location.href = `https://sandbox-secure.xsolla.com/paystation4/?token=${data.token}`
    } catch {
      alert('Checkout failed')
    }
  }

  return (
    <div className="store-page">

      {/* Logout Notification */}
      {logoutNotification && (
        <div className="logout-notification" role="alert">
          You have successfully logged out!
        </div>
      )}

      {/* Payment Success Modal */}
      {paymentSuccess && (
        <div className="payment-modal-overlay">
          <div className="payment-modal" role="dialog" aria-modal="true">
            <div className="payment-modal__icon">✅</div>
            <h2 className="payment-modal__title">Payment Successful!</h2>
            <p className="payment-modal__msg">
              Your items have been added to your inventory.
              <br />
              Your cart has been cleared.
            </p>
            <button
              className="payment-modal__close"
              type="button"
              onClick={() => setPaymentSuccess(false)}
            >
              Continue Shopping
            </button>
          </div>
        </div>
      )}

      {/* ── HEADER ── */}
      <header className="store-header">
        <div className="store-header__user">
          <div className="store-header__badge">
            <span className="store-header__label">LOGGED IN AS</span>
            <span className="store-header__email">{userEmail}</span>
          </div>
        </div>

        <div className="store-header__actions">
          <button
            className="icon-button"
            type="button"
            title="My Inventory"
            onClick={() => { setActiveTab('inventory'); setIsCartOpen(false) }}
          >
            🎒
          </button>

          <button
            className="cart-button"
            type="button"
            title="Cart"
            onClick={() => setIsCartOpen(true)}
          >
            🛒
            {cartCount > 0 && (
              <span className="cart-button__count">{cartCount}</span>
            )}
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
                    <img src={item.image_url} alt={item.name} className="item-card__image" />
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
                  <div className="quantity-badge">x{item.quantity}</div>
                  <div className="item-card__image-wrap">
                    <img src={item.image_url} alt={item.name} className="item-card__image" />
                  </div>
                  <div className="item-card__content">
                    <h3 className="item-card__name">{item.name}</h3>
                    <p className="item-card__description">{item.description}</p>
                  </div>
                </article>
              ))
            )}
          </section>
        )}
      </main>

      {/* ── CART DRAWER ── */}
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
                    <img src={item.image_url} alt={item.name} className="cart-item__image" />
                    <div className="cart-item__info">
                      <h4 className="cart-item__name">{item.name}</h4>
                      <p className="cart-item__price">
                        {item.currency} {Number(item.price).toFixed(2)}
                      </p>
                      <div className="cart-item__actions">
                        <button type="button" onClick={() => decreaseQuantity(item.id)}>-</button>
                        <span>{item.quantity}</span>
                        <button
                          type="button"
                          onClick={() => increaseQuantity(item.id)}
                          disabled={item.sku === LIMITED_SKU}
                          style={item.sku === LIMITED_SKU ? { opacity: 0.3, cursor: 'not-allowed' } : undefined}
                        >+</button>
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

      {/* ── FOOTER ── */}
      <footer className="store-footer">
        © 2024 Garden War Dashboard | All Rights Reserved
      </footer>
    </div>
  )
}
