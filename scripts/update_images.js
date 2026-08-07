/**
 * Updates image_url for already-created products, matching each product name
 * to a locally-available generic image (since product-specific images don't exist).
 *
 * USAGE:
 *   node update_images.js --token YOUR_ADMIN_TOKEN --images "C:\Users\Administrator\Desktop\Blinkit\assets"
 */

const fs = require('fs');
const path = require('path');

const BASE_URL = 'https://ecommerce-backend-dd4u.onrender.com/api/v1';

// Map: product name (must match exactly what's in the DB) -> preferred image filename (searched recursively)
const PRODUCT_IMAGE_MAP = {
  'Aashirvaad Atta': 'aata1.png',
  'Basmati Rice': 'rice1.png',
  'Toor Dal': 'daal1.png',
  'Moong Dal': 'daal2.png',
  'Chana Dal': 'daal3.png',
  'Sugar': 'sugar1.png',
  'Sunflower Oil': 'oil.png',
  'Mustard Oil': 'oil_ghee_masala.png',
  'Ghee': 'goverdhanghee.png',
  'Turmeric Powder': 'masala1.png',
  'Red Chilli Powder': 'masala2.png',
  'Garam Masala': 'garammasala.png',
  'Amul Milk': 'milk1.png',
  'Farm Eggs': 'egg1.png',
  'Bread': 'bread1.png',
  'Butter': 'butter1.png',
  'Paneer': 'butter2.png',
  'Cheese Slices': 'butter3.png',
  'Almonds': 'dryfruit1.png',
  'Cashews': 'dryfruit2.png',
  'Raisins': 'dryfruit3.png',
  'Dates': 'dryfruit4.png',
  'Walnuts': 'dryfruit5.png',
  'Non-Stick Tawa': 'cookware1.png',
  'Steel Water Bottle': 'bottle1.png',
  'Plastic Container Set': 'container1.png',
  'Kitchen Scissors': 'cookware2.png',
  'Chopping Board': 'cookware3.png',
  'Chicken Breast': 'chicken1.png',
  'Chicken Curry Cut': 'chicken2.png',
  'Mutton Curry Cut': 'mutton1.png',
  'Rohu Fish': 'fish1.png',
  'Prawns': 'seafood1.png',
  'Tata Tea': 'tea1.png',
  'Nescafe Coffee': 'coffee1.png',
  'Green Tea': 'greentea1.png',
  'Bournvita': 'malt1.png',
  'Horlicks': 'malt2.png',
  'Maggi Noodles': 'instant1.png',
  'Instant Pasta': 'instant2.png',
  'Ready to Eat Pulao': 'instant3.png',
  'Instant Soup': 'soup1.png',
  'Frozen Paratha': 'frozen1.png',
  'Elaichi': 'sauf1.png',
  'Saunf': 'sauf2.png',
  'Pan Masala': 'sauf3.png',
  'Mints': 'mint1.png',
  'Supari': 'supari1.png',
  'Face Wash': 'facewash1.png',
  'Moisturizer': 'moisturizer1.png',
  'Sunscreen SPF50': 'sunscreen1.png',
  'Face Mask': 'facemask1.png',
  'Lip Balm': 'moisturizer2.png',
  'Sanitary Pads': 'pad1.png',
  'Intimate Wash': 'intimate1.png',
  'Panty Liners': 'pad2.png',
  'Menstrual Cup': 'menstrualcup1.png',
  'Diapers': 'diaper1.png',
  'Baby Wipes': 'babywipes1.png',
  'Baby Lotion': 'babyskincare1.png',
  'Baby Shampoo': 'babyskincare2.png',
  'Baby Food': 'babyfood.png',
  'Paracetamol': 'medi2.png',
  'Multivitamin': 'medi3.png',
  'First Aid Kit': 'firstaid1.png',
  'Protein Powder': 'protein1.png',
  'Hand Sanitizer': 'firstaid2.png',
  'Floor Cleaner': 'floorcleanser1.png',
  'Detergent Powder': 'detergent1.png',
  'Dishwash Liquid': 'dishwash1.png',
  'Mosquito Repellent': 'Mosquitorepellents1.png',
  'Toilet Cleaner': 'floorcleanser2.png',
  'Earphones': 'earphone1.png',
  'USB Cable': 'charger1.png',
  'Power Bank': 'charger2.png',
  'LED Bulb': 'bulb1.png',
  'Mobile Charger': 'charger3.png',
};

function parseArgs() {
  const args = process.argv.slice(2);
  const out = {};
  for (let i = 0; i < args.length; i++) {
    if (args[i] === '--token') out.token = args[++i];
    if (args[i] === '--images') out.imagesDir = args[++i];
  }
  return out;
}

// Recursively find a file by exact name under a root dir. Returns first match or null.
function findFileRecursive(rootDir, targetName) {
  const stack = [rootDir];
  while (stack.length) {
    const dir = stack.pop();
    let entries;
    try {
      entries = fs.readdirSync(dir, { withFileTypes: true });
    } catch (e) {
      continue;
    }
    for (const entry of entries) {
      const full = path.join(dir, entry.name);
      if (entry.isDirectory()) {
        stack.push(full);
      } else if (entry.name.toLowerCase() === targetName.toLowerCase()) {
        return full;
      }
    }
  }
  return null;
}

async function fetchAllProducts(headers) {
  const all = [];
  let page = 1;
  while (true) {
    const res = await fetch(`${BASE_URL}/products/?limit=100&page=${page}`, { headers });
    const data = await res.json();
    const items = data.products || data.items || data.data || [];
    if (!Array.isArray(items) || items.length === 0) break;
    all.push(...items);
    if (items.length < 100) break;
    page++;
    if (page > 20) break; // safety cap
  }
  return all;
}

async function uploadImage(headers, localImagePath) {
  const fileBuffer = fs.readFileSync(localImagePath);
  const ext = path.extname(localImagePath).toLowerCase();
  const mime = ext === '.jpg' || ext === '.jpeg' ? 'image/jpeg' : ext === '.webp' ? 'image/webp' : 'image/png';
  const blob = new Blob([fileBuffer], { type: mime });
  const form = new FormData();
  form.append('image', blob, path.basename(localImagePath));

  const uploadRes = await fetch(`${BASE_URL}/upload`, {
    method: 'POST',
    headers,
    body: form,
  });
  const uploadData = await uploadRes.json();
  if (!uploadRes.ok) {
    throw new Error(uploadData.error || `upload failed: ${uploadRes.status}`);
  }
  return uploadData.url || uploadData.image_url || '';
}

async function updateProductImage(headers, product, imageUrl) {
  // Backend validator requires the full ProductRequest body (name, price, category_id all
  // required), so we resend the product's existing fields alongside the new image_url.
  const body = {
    name: product.name,
    description: product.description || `${product.name}`,
    price: product.price,
    image_url: imageUrl,
    category_id: product.category_id ?? product.categoryId ?? product.category?.id,
    stock: product.stock ?? 50,
  };

  // Try PUT first (more conventional for full-resource update), then PATCH as fallback.
  let res = await fetch(`${BASE_URL}/admin/products/${product.id}`, {
    method: 'PUT',
    headers: { ...headers, 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  });
  if (res.status === 404 || res.status === 405) {
    res = await fetch(`${BASE_URL}/admin/products/${product.id}`, {
      method: 'PATCH',
      headers: { ...headers, 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
    });
  }
  const data = await res.json().catch(() => ({}));
  if (!res.ok) {
    throw new Error(data.error || JSON.stringify(data) || `update failed: ${res.status}`);
  }
  return data;
}

async function main() {
  const { token, imagesDir } = parseArgs();
  if (!token || !imagesDir) {
    console.error('Usage: node update_images.js --token YOUR_ADMIN_TOKEN --images "path\\to\\Blinkit\\assets"');
    process.exit(1);
  }
  if (!fs.existsSync(imagesDir)) {
    console.error(`Images folder not found: ${imagesDir}`);
    process.exit(1);
  }

  const headers = { Authorization: `Bearer ${token}` };

  console.log('Fetching existing products...');
  const products = await fetchAllProducts(headers);
  console.log(`Found ${products.length} products in backend.`);
  if (products.length > 0) {
    console.log('Sample product shape (for debugging field names):', JSON.stringify(products[0]));
  }

  let ok = 0, fail = 0, notMapped = 0, imageMissing = 0;

  for (const [name, imageFile] of Object.entries(PRODUCT_IMAGE_MAP)) {
    const product = products.find((p) => p.name === name);
    if (!product) {
      console.error(`  SKIP ${name}: not found in backend product list`);
      notMapped++;
      continue;
    }

    const localPath = findFileRecursive(imagesDir, imageFile);
    if (!localPath) {
      console.error(`  FAIL ${name}: local image "${imageFile}" not found anywhere under ${imagesDir}`);
      imageMissing++;
      continue;
    }

    try {
      const imageUrl = await uploadImage(headers, localPath);
      await updateProductImage(headers, product, imageUrl);
      console.log(`  OK  ${name} -> ${imageFile}`);
      ok++;
    } catch (e) {
      console.error(`  FAIL ${name}:`, e.message);
      fail++;
    }
  }

  console.log(`\nDone. Updated: ${ok}, Failed: ${fail}, Not found in backend: ${notMapped}, Local image missing: ${imageMissing}`);
}

main();
