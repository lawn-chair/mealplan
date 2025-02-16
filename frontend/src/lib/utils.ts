import type { MealData, RecipeData } from '$lib/types';
import { PutObjectCommand, S3Client } from "@aws-sdk/client-s3";

export async function parseFormValues(submit : FormData) : Promise<RecipeData> {
    const name = submit.get('name');
    const description = submit.get('description');
    const slug = submit.get('slug');
    const ingredients: { amount: string; name: string; calories?: number }[] = [];
    const steps: { text: string; order: number; }[] = [];
    
    let stepCount = 0;
    const stepMap : Record <string, number> = {};

    for(const key of submit.keys()) {
        const matches = key.match(/ingredient\.(?<index>\d+)\.(?<key>\w+)/);
        if(matches && matches.groups) {
            const value = submit.get(key);
            const index = parseInt(matches.groups['index']);

            if(!ingredients[index]) {
                ingredients.push({amount: '', name: ''});
            }
            // @ts-expect-error using dynamic key
            ingredients[index][matches.groups['key']] = value;

            console.log(index + '/' + matches.groups.key + '/' + value);
        }

        const stepMatch = key.match(/step\.(?<index>\d+)\.(?<key>\w+)/);
        if(stepMatch && stepMatch.groups) {
            const value = submit.get(key);
            const index = stepMatch.groups['index'];

            if(stepMap[index] === undefined) {
                steps.push({text: '', order: 0});
                stepMap[index] = stepCount;
                stepCount++;
            }
            if(stepMatch.groups['key'] == 'order') {
                // @ts-expect-error value won't be null
                steps[stepMap[index]]['order'] = parseInt(value);
            } else {
                // @ts-expect-error using dynamic key
                steps[stepMap[index]][stepMatch.groups['key']] = value;
            } 
        }
        
    }

    let image = submit.get('image')?.toString() || null;

    const imageFile = submit.get('image_file') as File | null;
    console.log(imageFile);
    if (imageFile !== null && imageFile.size > 0) {
        // Create an instance of the S3Client
        const s3 = new S3Client({
            endpoint: 'http://localhost:9000',
            region: 'auto',
            credentials: {
                accessKeyId: 'zpFzVDPtycmDm3YegCmq',
                secretAccessKey: 'HVpXPcB4UQzFLKF0HdzcZYnLXIUmAN0aXceg4jtW',
            },
            tls: false,
            forcePathStyle: true, // needed with minio?
        });

        // Set the S3 bucket name and key for the uploaded file
        const bucketName = 'mp-images';
        const key = imageFile.name;
        // Create the parameters for the S3 upload
        
        const params = {
            Bucket: bucketName,
            Key: `${crypto.randomUUID()}.${key}`,
            Body: Buffer.from(await imageFile?.arrayBuffer()),
        };

        // Upload the file to the S3 bucket
        const response = await s3.send(new PutObjectCommand(params));
        if(response.$metadata.httpStatusCode === 200) {
            image = `http://localhost:9000/${bucketName}/${params.Key}`;
        }
    }

    return {
        name: name?.toString() || '',
        description: description?.toString() || '',
        slug: slug?.toString() || '',
        ingredients: ingredients || [],
        steps: steps || [],
        image: {Valid: image && image.length !== 0 ? true : false, String: image || ''}
    }
} 

export async function parseMealFormValues(submit : FormData) : Promise<MealData> {
    const data = await parseFormValues(submit);
    const recipes = [];

    for(const key of submit.keys()) {
        if(key.startsWith('recipe')) {
            const value = submit.get(key);
            if(value !== null) {
                recipes.push({recipe_id: parseInt(value.toString())});
            }
        }
    }
    
    return {
        ...data,
        recipes: recipes,
    };
}

export const updatePosition = (arr, oPos, nPos, offset = 0) => {
    if (typeof nPos === "number" && typeof oPos === "number" && nPos !== oPos) {
      arr.splice(nPos + offset, 0, arr.splice(oPos + offset, 1)[0])
    }
  
    arr.forEach((item, index) => {
      item.order = index + 1
    })
  
    console.log("Positioned: ", arr);
  
    return arr
  }