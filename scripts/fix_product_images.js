/**
 * Fixes the product images added by add_products_all_categories.js.
 * The previous script used random picsum.photos images (unrelated to the
 * product). This script fetches all products, finds the ones we added,
 * and updates each with a relevant image based on a keyword, using
 * loremflickr.com which returns a real photo matching the given tag.
 *
 * USAGE (from ecommerce-backend/scripts folder):
 *   node fix_product_images.js --token YOUR_ADMIN_TOKEN --base http://192.168.1.12:8081/api/v1
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

// product name -> search keyword for a relevant photo
const KEYWORDS = {
  'Brown Bread': 'bread',
  'Butter Croissant': 'croissant',
  'Orange Juice 1L': 'orange-juice',
  'Cola 750ml': 'cola-bottle',
  'Marie Gold Biscuits': 'biscuits',
  'Cream Biscuits': 'biscuits',
  'Dark Chocolate Bar': 'dark-chocolate',
  'Milk Chocolate Bar': 'chocolate-bar',
  'Wireless Earbuds': 'earbuds',
  'Power Bank 10000mAh': 'powerbank',
  'Smart Watch': 'smartwatch',
  'USB-C Cable': 'usb-cable',
  'Fresh Apples 1kg': 'apples',
  'Bananas 1 dozen': 'bananas',
  'Vanilla Ice Cream 500ml': 'vanilla-icecream',
  'Chocolate Ice Cream 500ml': 'chocolate-icecream',
  'Tomato Ketchup 500g': 'ketchup',
  'Chilli Ketchup 350g': 'chilli-sauce',
  'Aloo Bhujia 200g': 'indian-snacks',
  'Mixture 200g': 'namkeen',
  'Hand Sanitizer 100ml': 'hand-sanitizer',
  'Face Wash 100ml': 'face-wash',
  'Mango Pickle 400g': 'mango-pickle',
  'Mixed Veg Pickle 400g': 'pickle-jar',
  'Agarbatti Pack': 'incense-sticks',
  'Diya Set (Pack of 12)': 'diya-lamp',
  'Anti-Dandruff Shampoo 200ml': 'shampoo-bottle',
  'Herbal Shampoo 340ml': 'shampoo-bottle',
  'Potato Chips 100g': 'potato-chips',
  'Peanut Snacks 150g': 'peanuts',
  'Sandalwood Soap': 'soap-bar',
  'Neem Soap': 'soap-bar',
  'Building Blocks Set': 'building-blocks-toy',
  'Remote Control Car': 'rc-car-toy',
  'Chocolate Wafers 100g': 'wafer-biscuit',
  'Cream Wafers 100g': 'wafer-biscuit',
};

async function main() {
  const { token, base } = parseArgs();
  if (!token) {
    console.error('Usage: node fix_product_images.js --token YOUR_ADMIN_TOKEN [--base http://host:port/api/v1]');
    process.exit(1);
  }

  const headers = {
    Authorization: `Bearer ${token}`,
    'Content-Type': 'application/json',
  };

  console.log('Fetching all products...');
  let all = [];
  let page = 1;
  const limit = 200;
  let consecutiveEmpty = 0;
  while (page <= 100) { // safety cap
    const res = await fetch(`${base}/products?page=${page}&limit=${limit}`);
    const data = await res.json();
    const batch = data.products || [];
    if (batch.length === 0) {
      break; // no more pages
    }
    all = all.concat(batch);
    console.log(`  fetched page ${page}, batch size ${batch.length}, total so far ${all.length}`);
    page++;
  }
  console.log(`Fetched ${all.length} products total.\n`);

  let ok = 0, fail = 0, skipped = 0;
  for (const p of all) {
    const keyword = KEYWORDS[p.name];
    if (!keyword) {
      skipped++;
      continue;
    }
    const imageUrl = `https://loremflickr.com/400/400/${keyword}`;
    const body = {
      name: p.name,
      description: p.description,
      price: p.price,
      image_url: imageUrl,
      category_id: p.category_id,
      stock: p.stock,
    };
    try {
      const res = await fetch(`${base}/admin/products/${p.id}`, {
        method: 'PUT',
        headers,
        body: JSON.stringify(body),
      });
      if (res.ok) {
        ok++;
        console.log(`  FIXED id ${p.id}: ${p.name} -> ${imageUrl}`);
      } else {
        fail++;
        const data = await res.json().catch(() => ({}));
        console.error(`  FAIL id ${p.id} (${p.name}):`, data.error || res.status);
      }
    } catch (e) {
      fail++;
      console.error(`  FAIL id ${p.id} (${p.name}):`, e.message);
    }
  }

  console.log(`\nDone. Fixed: ${ok}, Failed: ${fail}, Skipped (not in list): ${skipped}`);
}

main().catch((e) => {
  console.error('Fatal error:', e);
  process.exit(1);
});
