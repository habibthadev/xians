import { API_URL } from '$env/static/private';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ params, fetch, setHeaders }) => {
  const apiUrl = API_URL || 'http://localhost:8080';
  const slug = params.slug;

  try {
    const res = await fetch(`${apiUrl}/api/g/${slug}`, {
      headers: { Accept: 'application/json' },
    });

    if (!res.ok) {
      if (res.status === 401) {
        return {
          slug,
          requiresPassword: true,
          title: 'Password-Protected Gallery - xians',
          description: 'This gallery requires a password to view.',
        };
      }
      return {
        slug,
        requiresPassword: false,
        title: 'Gallery Not Found - xians',
        description: 'This gallery does not exist or has expired.',
        status: res.status,
      };
    }

    const data = await res.json();

    const fileCount = data.files?.length || 0;
    const title = `Gallery - xians`;
    const description = `${fileCount} file${fileCount !== 1 ? 's' : ''} shared via xians`;

    setHeaders({
      'Cache-Control': 'public, max-age=60, s-maxage=60',
    });

    return {
      slug,
      requiresPassword: false,
      title,
      description,
      files: data.files || [],
      createdAt: data.createdAt,
      expiresAt: data.expiresAt,
      downloadCount: data.downloadCount,
      maxDownloads: data.maxDownloads,
    };
  } catch {
    return {
      slug,
      requiresPassword: false,
      title: 'Gallery - xians',
      description: 'Shared media gallery',
    };
  }
};
