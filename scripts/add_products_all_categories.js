/**
 * Adds sample products across all your app's categories (2 products each),
 * each with its own unique image, using the admin Products API.
 *
 * USAGE (from ecommerce-backend/scripts folder):
 *   node add_products_all_categories.js --token YOUR_ADMIN_TOKEN --base http://192.168.1.12:8081/api/v1
 *
 * --base is optional, defaults to http://192.168.1.12:8081/api/v1
 */

function parseArgs() {
  const args = process.argv.slice(2);
  const out = { base: 'http://192.168.1.12:8081/api/v1' };
  for (let i = 0; i < args.length; i++) {
    if (args[i] === '--token') out.token = args[++i];
    if (args[i] === '--base') out.base = args[++i];
  }
  return out;
}

// category_id -> [ { name, price, description } ]
const CATALOG = {
  12: [ // Bakery
    { name: 'Brown Bread', price: 45, description: 'Fresh whole wheat bread' },
    { name: 'Butter Croissant', price: 60, description: 'Flaky butter croissant' },
  ],
  10: [ // Beverages
    { name: 'Orange Juice 1L', price: 120, description: 'Fresh orange juice' },
    { name: 'Cola 750ml', price: 50, description: 'Chilled cola drink' },
  ],
  13: [ // Biscuits
    { name: 'Marie Gold Biscuits', price: 30, description: 'Classic tea-time biscuits' },
    { name: 'Cream Biscuits', price: 35, description: 'Chocolate cream biscuits' },
  ],
  9: [ // Chocolate
    { name: 'Dark Chocolate Bar', price: 90, description: '70% cocoa dark chocolate' },
    { name: 'Milk Chocolate Bar', price: 80, description: 'Creamy milk chocolate' },
  ],
  1: [ // Electronics
    { name: 'Wireless Earbuds', price: 1499, description: 'Bluetooth 5.0 earbuds' },
    { name: 'Power Bank 10000mAh', price: 999, description: 'Fast charging power bank' },
  ],
  4: [ // Electronics & Gadgets
    { name: 'Smart Watch', price: 2499, description: 'Fitness tracking smart watch' },
    { name: 'USB-C Cable', price: 199, description: 'Fast charging cable, 1m' },
  ],
  8: [ // Fruits
    { name: 'Fresh Apples 1kg', price: 150, description: 'Farm fresh apples' },
    { name: 'Bananas 1 dozen', price: 60, description: 'Ripe yellow bananas' },
  ],
  11: [ // Ice Creams
    { name: 'Vanilla Ice Cream 500ml', price: 130, description: 'Classic vanilla ice cream' },
    { name: 'Chocolate Ice Cream 500ml', price: 140, description: 'Rich chocolate ice cream' },
  ],
  16: [ // Ketchup
    { name: 'Tomato Ketchup 500g', price: 90, description: 'Thick tomato ketchup' },
    { name: 'Chilli Ketchup 350g', price: 85, description: 'Spicy chilli ketchup' },
  ],
  14: [ // Namkeen
    { name: 'Aloo Bhujia 200g', price: 40, description: 'Crispy potato namkeen' },
    { name: 'Mixture 200g', price: 45, description: 'Spicy mixed namkeen' },
  ],
  19: [ // Personal Care
    { name: 'Hand Sanitizer 100ml', price: 60, description: 'Alcohol based sanitizer' },
    { name: 'Face Wash 100ml', price: 150, description: 'Gentle daily face wash' },
  ],
  20: [ // Pickle
    { name: 'Mango Pickle 400g', price: 110, description: 'Traditional mango pickle' },
    { name: 'Mixed Veg Pickle 400g', price: 100, description: 'Spicy mixed vegetable pickle' },
  ],
  21: [ // Puja Items
    { name: 'Agarbatti Pack', price: 50, description: 'Fragrant incense sticks' },
    { name: 'Diya Set (Pack of 12)', price: 80, description: 'Clay oil lamps for puja' },
  ],
  17: [ // Shampoo
    { name: 'Anti-Dandruff Shampoo 200ml', price: 180, description: 'Clears dandruff, nourishes scalp' },
    { name: 'Herbal Shampoo 340ml', price: 220, description: 'Natural herbal shampoo' },
  ],
  2: [ // Snacks
    { name: 'Potato Chips 100g', price: 30, description: 'Crispy salted chips' },
    { name: 'Peanut Snacks 150g', price: 40, description: 'Roasted spicy peanuts' },
  ],
  18: [ // Soap
    { name: 'Sandalwood Soap', price: 40, description: 'Moisturizing sandalwood soap' },
    { name: 'Neem Soap', price: 35, description: 'Antibacterial neem soap' },
  ],
  22: [ // Toys
    { name: 'Building Blocks Set', price: 499, description: '100-piece building blocks' },
    { name: 'Remote Control Car', price: 899, description: 'Fast RC car for kids' },
  ],
  15: [ // Wafers
    { name: 'Chocolate Wafers 100g', price: 35, description: 'Crispy chocolate wafers' },
    { name: 'Cream Wafers 100g', price: 35, description: 'Vanilla cream wafers' },
  ],
};

function slugify(name) {
  return name.toLowerCase().replace(/[^a-z0-9]+/g, '-').replace(/(^-|-$)/g, '');
}

async function main() {
  const { token, base } = parseArgs();
  if (!token) {
    console.error('Usage: node add_products_all_categories.js --token YOUR_ADMIN_TOKEN [--base http://host:port/api/v1]');
    process.exit(1);
  }

  const headers = {
    Authorization: `Bearer ${token}`,
    'Content-Type': 'application/json',
  };

  let ok = 0, fail = 0;

  for (const [categoryId, products] of Object.entries(CATALOG)) {
    for (const p of products) {
      const imageUrl = `https://picsum.photos/seed/${slugify(p.name)}/400/400`;
      const body = {
        name: p.name,
        description: p.description,
        price: p.price,
        image_url: imageUrl,
        category_id: Number(categoryId),
        stock: 50,
      };
      try {
        const res = await fetch(`${base}/admin/products`, {
          method: 'POST',
          headers,
          body: JSON.stringify(body),
        });
        if (res.ok) {
          ok++;
          console.log(`  ADDED: ${p.name} (category ${categoryId})`);
        } else {
          fail++;
          const data = await res.json().catch(() => ({}));
          console.error(`  FAIL: ${p.name} ->`, data.error || res.status);
        }
      } catch (e) {
        fail++;
        console.error(`  FAIL: ${p.name} ->`, e.message);
      }
    }
  }

  console.log(`\nDone. Added: ${ok}, Failed: ${fail}`);
}

main().catch((e) => {
  console.error('Fatal error:', e);
  process.exit(1);
});
