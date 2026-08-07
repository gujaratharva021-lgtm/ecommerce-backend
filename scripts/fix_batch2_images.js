/**
 * Fixes the 74 "batch 2" products (Atta & Rice, Dairy & Eggs, Electronics,
 * etc.) that had no image, now that we know real photos for them exist as
 * flat files under assets/images/products/ (e.g. "rice1.png",
 * "earphone1.png") rather than in per-category subfolders.
 *
 * Uploads the matched local file to Cloudinary and PATCHes the existing
 * product's image_url.
 *
 * USAGE:
 *   node fix_batch2_images.js --token YOUR_ADMIN_TOKEN --images "C:\Users\Administrator\Desktop\Blinkit\assets"
 */

const fs = require('fs');
const path = require('path');

const BASE_URL = 'https://ecommerce-backend-dd4u.onrender.com/api/v1';

// Product name (from blinkit_products_batch2.json) -> filename inside assets/images/products/
const NAME_TO_FILE = {
  'Aashirvaad Atta': 'aata1.png',
  'Basmati Rice': 'rice1.png',
  'Toor Dal': 'daal1.png',
  'Moong Dal': 'daal2.png',
  'Chana Dal': 'daal3.png',
  'Sugar': 'sugar1.png',

  'Sunflower Oil': 'oil.png',
  'Mustard Oil': 'saffola.png',
  'Ghee': 'goverdhanghee.png',
  'Turmeric Powder': 'masala1.png',
  'Red Chilli Powder': 'masala2.png',
  'Garam Masala': 'mdhmasala.png',

  'Amul Milk': 'milk1.png',
  'Farm Eggs': 'egg1.png',
  'Bread': 'bread1.png',
  'Butter': 'butter1.png',
  'Paneer': 'milk3.png',
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
  'Horlicks': 'healthdrinks1.png',

  'Maggi Noodles': 'instant1.png',
  'Instant Pasta': 'instant2.png',
  'Ready to Eat Pulao': 'instant3.png',
  'Instant Soup': 'soup1.png',
  'Frozen Paratha': 'frozen1.png',

  'Elaichi': 'mint1.png',
  'Saunf': 'sauf1.png',
  'Pan Masala': 'mint2.png',
  'Mints': 'mint3.png',
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
  'Multivitamin': 'protein2.png',
  'First Aid Kit': 'firstaid1.png',
  'Protein Powder': 'protein1.png',
  'Hand Sanitizer': 'medi3.png',

  'Floor Cleaner': 'floorcleanser1.png',
  'Detergent Powder': 'detergent1.png',
  'Dishwash Liquid': 'dishwash1.png',
  'Mosquito Repellent': 'Mosquitorepellents1.png',
  'Toilet Cleaner': 'floorcleanser2.png',

  'Earphones': 'earphone1.png',
  'USB Cable': 'charger2.png',
  'Power Bank': 'batteries1.png',
  'LED Bulb': 'bulb1.png',
  'Mobile Charger': 'charger1.png',
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

async function fetchAllProducts(headers) {
  const all = [];
  let page = 1;
  while (true) {
    const res = await fetch(`${BASE_URL}/products?limit=100&page=${page}`);
    const data = await res.json();
    all.push(...(data.products || []));
    if (page >= (data.total_pages || 1)) break;
    page++;
  }
  return all;
}

async function main() {
  const { token, imagesDir } = parseArgs();
  if (!token || !imagesDir) {
    console.error('Usage: node fix_batch2_images.js --token YOUR_ADMIN_TOKEN --images "path\\to\\Blinkit\\assets"');
    process.exit(1);
  }

  const headers = { Authorization: `Bearer ${token}` };

  console.log('Fetching existing products from backend...');
  const products = await fetchAllProducts(headers);
  console.log(`Found ${products.length} products in backend.`);

  let updated = 0, skipped = 0, failed = 0;

  for (const product of products) {
    const filename = NAME_TO_FILE[product.name];
    if (!filename) {
      skipped++;
      continue;
    }
    const localImagePath = path.join(imagesDir, 'images', 'products', filename);
    if (!fs.existsSync(localImagePath)) {
      console.error(`  FAIL ${product.name}: expected file not found at ${localImagePath}`);
      failed++;
      continue;
    }

    try {
      const fileBuffer = fs.readFileSync(localImagePath);
      const ext = path.extname(localImagePath).toLowerCase();
      const mime = ext === '.jpg' || ext === '.jpeg' ? 'image/jpeg' : ext === '.webp' ? 'image/webp' : 'image/png';
      const blob = new Blob([fileBuffer], { type: mime });
      const form = new FormData();
      form.append('image', blob, filename);

      const uploadRes = await fetch(`${BASE_URL}/upload`, {
        method: 'POST',
        headers,
        body: form,
      });
      const uploadData = await uploadRes.json();
      if (!uploadRes.ok || !uploadData.image_url) {
        console.error(`  FAIL ${product.name}: upload failed - ${uploadData.error || uploadRes.status}`);
        failed++;
        continue;
      }

      const putRes = await fetch(`${BASE_URL}/admin/products/${product.id}`, {
        method: 'PUT',
        headers: { ...headers, 'Content-Type': 'application/json' },
        body: JSON.stringify({
          name: product.name,
          description: product.description,
          price: product.price,
          image_url: uploadData.image_url,
          category_id: product.category_id,
        }),
      });
      const putData = await putRes.json();
      if (putRes.ok) {
        updated++;
        console.log(`  OK  ${product.name} -> ${uploadData.image_url}`);
      } else {
        failed++;
        console.error(`  FAIL ${product.name}: update failed - ${putData.error || putRes.status}`);
      }
    } catch (e) {
      failed++;
      console.error(`  FAIL ${product.name}: ${e.message}`);
    }
  }

  console.log(`\nDone. Updated: ${updated}, Failed: ${failed}, Skipped (not in mapping): ${skipped}`);
}

main();