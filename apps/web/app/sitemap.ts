import { MetadataRoute } from 'next';
import { navigationConfig, type NavItem } from '@/lib/navigation-config';

export default function sitemap(): MetadataRoute.Sitemap {
  const baseUrl = 'https://salesview.id';
  const locales = ['id', 'en'] as const;
  const marketingPaths = [
    '/',
    '/pricing',
    '/accounting',
    '/invoicing',
    '/fixed-assets',
    '/financial-reports',
    '/reconciliation',
    '/crm',
    '/sales',
    '/quotations',
    '/purchase',
    '/stock',
    '/goods-receipt',
    '/movements',
    '/employees',
    '/attendance',
    '/recruitment',
    '/travel-planner',
    '/evaluation',
    '/pos',
    '/about',
    '/templates',
  ] as const;
  
  const appsMenus = navigationConfig.filter(
    (nav) => nav.children && nav.name !== 'Reports' && nav.name !== 'Master Data' && nav.name !== 'Dashboard'
  );

  const getSlug = (menu: NavItem) => {
    const firstChild = menu.children?.find((c) => c.url);
    return firstChild ? firstChild.url.split('/').filter(Boolean)[0] || 'app' : 'app';
  };

  const appUrlsId = appsMenus.map((menu) => ({
    url: `${baseUrl}/id/apps/${getSlug(menu)}`,
    lastModified: new Date(),
    changeFrequency: 'weekly' as const,
    priority: 0.8,
  }));

  const appUrlsEn = appsMenus.map((menu) => ({
    url: `${baseUrl}/en/apps/${getSlug(menu)}`,
    lastModified: new Date(),
    changeFrequency: 'weekly' as const,
    priority: 0.8,
  }));

  const marketingUrls = locales.flatMap((locale) =>
    marketingPaths.map((path) => {
      const localizedPath = path === '/' ? `/${locale}` : `/${locale}${path}`;
      const isCore = path === '/';
      const isPricing = path === '/pricing';

      return {
        url: `${baseUrl}${localizedPath}`,
        lastModified: new Date(),
        changeFrequency: (isCore ? 'monthly' : 'weekly') as 'monthly' | 'weekly',
        priority: isCore ? 1 : isPricing ? 0.95 : 0.9,
      };
    })
  );

  return [...marketingUrls, ...appUrlsId, ...appUrlsEn];
}
