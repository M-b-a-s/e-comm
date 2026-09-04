# Product Requirements Document: Audiophile E-Commerce Platform

**Source:** Figma file "audiophile-ecommerce-website (Copy)"
**Prepared for:** Backend engineering (two parallel implementations — REST and GraphQL)
**Status:** Draft v1.0

---

## 1. Overview

Audiophile is a single-brand e-commerce storefront selling premium audio equipment across three categories: **Headphones**, **Speakers**, and **Earphones**. The Figma file contains full responsive designs (Mobile 375px / Tablet 768px / Desktop 1440px) for every screen in the shopper journey: browse → product detail → cart → checkout → order confirmation.

This document translates those screens into product and API requirements. Since the goal is to build **two independent backends (REST and GraphQL)** against the same domain, the data model, business rules, and validation logic are specified backend-agnostically in §5–7, with concrete endpoint/schema sketches for each style in §8.

---

## 2. Goals & Non-Goals

**Goals**
- Support full catalog browsing (home, category, product detail) with related-product recommendations.
- Support a persistent shopping cart with quantity management.
- Support a single-page checkout that captures billing, shipping, and payment details and produces an order.
- Expose the same catalog/cart/order domain via both a REST API and a GraphQL API, so either can be swapped in behind the same frontend.

**Non-Goals**
- Real payment processing (design shows "e-Money" and "Cash on Delivery" as mocked options — no PCI-scope gateway integration is implied).
- User accounts, login, order history, or wishlists (no such screens exist in the file).
- Admin/CMS tooling for managing products (content appears static/seeded).
- Search, filtering, sorting, or reviews (not present in the designs).

---

## 3. Personas

| Persona | Description |
|---|---|
| **Shopper** | Anonymous visitor browsing on mobile, tablet, or desktop; can add items to cart and check out without an account. |
| **Store operator (implicit)** | Not represented in the UI, but needs a way to seed/manage the 6-product catalog. Assumption: managed via seed data/DB directly for v1. |

---

## 4. Information Architecture / Screens

All screens are designed at 3 breakpoints (Mobile 375 / Tablet 768 / Desktop 1440). Functionally identical across breakpoints; layout differs only.

| Screen | Purpose | Key elements |
|---|---|---|
| **Home** | Landing page | Hero (featured new product: XX99 Mark II), 3 category tiles (Headphones/Speakers/Earphones), promo blocks for ZX9, ZX7, YX1, "About" blurb, footer |
| **Category listing** (×3: Headphones, Speakers, Earphones) | List all products in one category | Product cards (image, name, "See Product" CTA), ordered newest-first |
| **Product detail** (×6: XX99 Mark II, XX99 Mark I, XX59, ZX9, ZX7, YX1) | Convert to cart add | Gallery (3 images), name, category, price, description, quantity stepper, "Add to Cart", Features (rich text), "In the Box" (qty + item list), 2nd gallery block, "You may also like" (3 cross-category recommendations), category grid, About blurb |
| **Cart (mini/overlay)** | Review & adjust items before checkout | Line items (image, name, price, qty stepper with +/-), "Remove all", Total, "Checkout" CTA, item count badge (e.g. "CART (3)") |
| **Checkout** | Capture order details | Billing (Name, Email — with inline validation e.g. "wrong format"), Shipping (Address, ZIP Code, City, Country), Payment Method (radio: e-Money / Cash on Delivery), conditional e-Money fields (Number, PIN), Order summary sidebar (items, Total, Shipping, VAT (included), Grand Total), "Continue & Pay" |
| **Checkout — success modal** | Order confirmation | "Thank you for your order", "You will receive an email confirmation shortly", order summary (first item + "and N other item(s)"), Grand Total, "Back to Home" |
| **Mobile menu** | Nav drawer (mobile/tablet only) | Home / Headphones / Speakers / Earphones links |

**Global nav** (all pages): logo, Home / Headphones / Speakers / Earphones links, cart icon (with badge), hamburger on mobile/tablet.

---

## 5. Domain Data Model

```
Category
├─ id
├─ slug            (headphones | speakers | earphones)
└─ name

Product
├─ id
├─ slug                     (url-friendly, e.g. "xx99-mark-two-headphones")
├─ name                     (e.g. "XX99 Mark II Headphones")
├─ shortName                (e.g. "XX99 Mark II" — used in cart/summary)
├─ categoryId → Category
├─ isNew                    (bool — drives "NEW PRODUCT" badge, e.g. XX99 Mark II)
├─ priceCents                (integer, e.g. 299900 for $2,999)
├─ description               (product page intro paragraph)
├─ features                  (long-form rich text)
├─ boxIncludes[]              { quantity: int, item: string }   e.g. "1x Headphone Unit"
├─ gallery[]                  (image refs — 3 images per product: 2 secondary + 1 large)
├─ categoryImage               (image used on the category card / "You may also like")
└─ recommendedProductIds[]      (curated cross-sell list, 3 per product; not auto-generated)

CartItem  (session/cart-scoped, not persisted long-term)
├─ productId → Product
├─ quantity          (int ≥ 1)
└─ unitPriceCents      (snapshot of price at add-time, optional but recommended)

Cart
├─ id / sessionId
├─ items: CartItem[]
├─ totalCents            (sum of item price × qty)
└─ itemCount

Order
├─ id
├─ orderNumber
├─ items: OrderItem[]      { productId, name, unitPriceCents, quantity }
├─ billing: { name, email }
├─ shipping: { address, city, zipCode, country }
├─ paymentMethod          (e-money | cash-on-delivery)
├─ paymentDetails          (eMoneyNumber, eMoneyPin — only if e-money; store masked/last-4 only, never raw PIN in prod)
├─ subtotalCents
├─ shippingCents            (flat $50 in designs)
├─ vatCents                  (shown as "VAT (included)" — i.e. VAT is embedded in item price, not added on top)
├─ grandTotalCents            (subtotal + shipping; VAT is informational, already included)
├─ status                     (pending | confirmed)
└─ createdAt
```

### Observed pricing example (for validation against seed data / tests)
From the checkout screens: Total `$5,396` (XX99 MK II ×1 @ $2,999 + XX59 ×2 @ $899 + YX1 ×1 @ $599 = $5,396) → Shipping `$50` → VAT (included) `$1,079` → Grand Total `$5,446` (= Total + Shipping; VAT is a sub-component of Total, not additive).

**⚠ Business rule to confirm with design/finance:** VAT is labeled "(included)", meaning it's already inside the item price, not added on top. Grand Total = Subtotal + Shipping. VAT is purely informational on the summary. Both backends must implement this identically.

### Product catalog (from design content)

| Product | Category | Price | New? |
|---|---|---|---|
| XX99 Mark II Headphones | Headphones | $2,999 | Yes |
| XX99 Mark I Headphones | Headphones | $1,750 | No |
| XX59 Headphones | Headphones | $899 | No |
| ZX9 Speaker | Speakers | $4,500 | No |
| ZX7 Speaker | Speakers | $3,500 | No |
| YX1 Earphones | Earphones | $599 | No |

---

## 6. Functional Requirements

### 6.1 Catalog
- FR-1: Retrieve the full category list (3 fixed categories).
- FR-2: Retrieve all products within a category, ordered as designed (newest/featured first).
- FR-3: Retrieve a single product by id/slug with full detail (description, features, box contents, gallery, recommended products).
- FR-4: Retrieve a product's "recommended" set (3 curated cross-category products) — curated, not computed.
- FR-5: Retrieve the single "featured/new" product for the homepage hero (currently XX99 Mark II).

### 6.2 Cart
- FR-6: Add a product (with quantity ≥ 1) to the cart.
- FR-7: Update quantity of a cart line item (increment/decrement via stepper; minimum quantity 1, no explicit max observed in design).
- FR-8: Remove a single item or clear all items ("Remove all").
- FR-9: Retrieve current cart contents with computed total and item count.
- FR-10: Cart must persist across the browsing session (cross-screen; the mini-cart badge shows a running count e.g. "CART (3)").

### 6.3 Checkout
- FR-11: Accept billing details (name, email) with server-side email format validation (design shows an explicit "Wrong format" inline error state).
- FR-12: Accept shipping details (address, ZIP/postal code, city, country) — all required.
- FR-13: Accept a payment method selection: `e-money` or `cash-on-delivery`.
  - If `e-money`: require e-Money Number and e-Money PIN.
  - If `cash-on-delivery`: no additional fields; UI shows explanatory copy about paying on delivery.
- FR-14: Compute order summary server-side (never trust client-submitted totals): subtotal from live cart, flat shipping fee, VAT-inclusive total, grand total.
- FR-15: On submit, create an Order, decrement/snapshot cart, and return confirmation data (order id, item summary — "first item + N other item(s)" pattern, grand total).
- FR-16: Confirmation screen must be able to render from the Order response alone (no additional fetches implied by the design).

### 6.4 Validation rules (from visible design states)
- Email: must match a standard email format (inline error observed).
- Required fields (visually indicated as required inputs): Name, Email, Address, ZIP, City, Country, Phone Number, and conditionally e-Money Number/PIN.
- Quantity stepper: cannot go below 1 (minus button shown disabled/dimmed at qty 1 in the design).

---

## 7. Business Rules Summary

| Rule | Value / Logic |
|---|---|
| Shipping fee | Flat $50 per order (not per item, not weight-based) |
| VAT | Included in item price; displayed as informational line, not added to total |
| Grand Total | `subtotal + shipping` |
| Currency | USD, whole-dollar increments in design (store as integer cents to avoid float errors) |
| Cart minimum qty | 1 (cannot decrement to 0 via stepper — use Remove instead) |
| Recommended products | Curated per-product list of exactly 3, not algorithmic |

---

## 8. API Requirements

Both backends must expose equivalent capability. Field names below are suggestions; keep them consistent across both implementations so the frontend contract is interchangeable.

### 8.1 REST API sketch

| Method | Path | Purpose |
|---|---|---|
| GET | `/categories` | List 3 categories |
| GET | `/products?category=headphones` | List products, optional category filter |
| GET | `/products/{slug}` | Product detail incl. gallery, features, box contents |
| GET | `/products/{slug}/recommended` | 3 curated recommended products |
| GET | `/products/featured` | Homepage hero product |
| GET | `/cart` | Current cart (session-based) |
| POST | `/cart/items` | Add item `{ productId, quantity }` |
| PATCH | `/cart/items/{productId}` | Update quantity `{ quantity }` |
| DELETE | `/cart/items/{productId}` | Remove single item |
| DELETE | `/cart/items` | Remove all |
| POST | `/checkout` | Submit order `{ billing, shipping, paymentMethod, paymentDetails? }`; server pulls cart server-side |
| GET | `/orders/{id}` | Fetch confirmation detail |

Standard REST conventions: JSON bodies, 4xx with field-level validation errors for checkout (to drive the inline "Wrong format" style UI), idempotent PATCH/DELETE, session/cart identified via cookie or bearer token.

### 8.2 GraphQL API sketch

```graphql
type Category { id: ID!, slug: String!, name: String! }

type Money { cents: Int!, formatted: String! }

type BoxItem { quantity: Int!, item: String! }

type Product {
  id: ID!
  slug: String!
  name: String!
  shortName: String!
  category: Category!
  isNew: Boolean!
  price: Money!
  description: String!
  features: String!
  boxIncludes: [BoxItem!]!
  gallery: [String!]!
  recommended: [Product!]!
}

type CartItem { product: Product!, quantity: Int! }

type Cart {
  id: ID!
  items: [CartItem!]!
  itemCount: Int!
  total: Money!
}

type OrderItem { product: Product!, quantity: Int!, unitPrice: Money! }

type Order {
  id: ID!
  orderNumber: String!
  items: [OrderItem!]!
  subtotal: Money!
  shipping: Money!
  vat: Money!
  grandTotal: Money!
  status: String!
}

type Query {
  categories: [Category!]!
  products(category: String): [Product!]!
  product(slug: String!): Product
  featuredProduct: Product!
  cart: Cart!
  order(id: ID!): Order
}

input AddCartItemInput { productId: ID!, quantity: Int! }
input UpdateCartItemInput { productId: ID!, quantity: Int! }
input BillingInput { name: String!, email: String! }
input ShippingInput { address: String!, city: String!, zipCode: String!, country: String!, phone: String! }
input CheckoutInput {
  billing: BillingInput!
  shipping: ShippingInput!
  paymentMethod: PaymentMethod!
  eMoneyNumber: String
  eMoneyPin: String
}
enum PaymentMethod { E_MONEY, CASH_ON_DELIVERY }

type Mutation {
  addCartItem(input: AddCartItemInput!): Cart!
  updateCartItem(input: UpdateCartItemInput!): Cart!
  removeCartItem(productId: ID!): Cart!
  clearCart: Cart!
  checkout(input: CheckoutInput!): Order!
}
```

### 8.3 Cross-cutting API requirements
- Both backends compute totals server-side from the persisted cart — never accept a client-supplied total.
- Both backends validate checkout inputs identically (email format, required fields, conditional e-money fields) and return structured, field-level errors.
- Both backends should be able to run against the same underlying database/seed data so responses are comparable during development.
- Prices are transmitted as integer cents (or a `Money` shape) to avoid floating-point rounding differences between the two implementations.

---

## 9. Non-Functional Requirements

- **Responsive parity:** API responses are breakpoint-agnostic; all layout logic (Mobile/Tablet/Desktop) stays in the frontend. No backend requirement here beyond providing complete data (e.g. full gallery set for however many images each breakpoint chooses to show).
- **Statelessness / session handling:** Cart must work for anonymous users — use a session cookie or client-generated cart token; no auth system implied by the designs.
- **Data integrity:** Product prices, box contents, and recommendations should be centrally seeded/versioned so REST and GraphQL stay in sync (single source of truth, e.g. a shared seed script or DB).
- **Security:** Never persist raw e-Money PIN; mask or tokenize payment detail fields even though this is a mocked payment flow.

---

## 10. Open Questions / Assumptions

1. **VAT interpretation** — confirmed from screen math as "included in price, informational only." Flag this explicitly to stakeholders since it's easy to misimplement as additive.
2. **Recommended products** — treated as curated/static per product (matches the fixed 3-item "You may also like" sets seen per product page), not a recommendation engine. Confirm before building anything algorithmic.
3. **No accounts/order history** — confirm this is intentionally out of scope for v1, since the design has no login or "my orders" screens.
4. **Cash on Delivery validation** — no extra fields are captured for this method beyond the shared billing/shipping block; confirm nothing else is needed server-side (e.g. delivery notes).
5. **Stock/inventory** — not represented anywhere in the UI; assume unlimited stock for v1 unless told otherwise.

---

## 11. Appendix: Screen Inventory (from Figma)

Frames found in the file, per breakpoint (Mobile / Tablet / Desktop): Menu, Checkout (+ Modal), Cart, Product Detail (Earphones ×1, Speakers ×2, Headphones ×3), Category (Earphones, Speakers, Headphones), Home. Plus two "Design System" reference pages (not user-facing screens — typography/color/spacing tokens for the frontend team).