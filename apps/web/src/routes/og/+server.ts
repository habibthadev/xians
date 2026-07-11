import satori from 'satori';
import { Resvg } from '@resvg/resvg-js';
import type { RequestHandler } from './$types';
import { z } from 'zod';

const querySchema = z.object({
  title: z.string().min(1).max(120).default('xians'),
  description: z.string().max(200).optional().default('Share files instantly — no sign-up, no fuss.'),
});

const geistBold = fetch(
  'https://cdn.jsdelivr.net/npm/geist@1.3.0/dist/fonts/geist-sans/Geist-Bold.ttf'
).then((r) => r.arrayBuffer());

const geistRegular = fetch(
  'https://cdn.jsdelivr.net/npm/geist@1.3.0/dist/fonts/geist-sans/Geist-Regular.ttf'
).then((r) => r.arrayBuffer());

const anton = fetch(
  'https://fonts.gstatic.com/s/anton/v27/1Ptgg87LROyAm0K0.ttf'
).then((r) => r.arrayBuffer());

const poppinsBold = fetch(
  'https://fonts.gstatic.com/s/poppins/v24/pxiByp8kv8JHgFVrLCz7V1s.ttf'
).then((r) => r.arrayBuffer());

interface VNode {
  type: string;
  props: Record<string, unknown>;
}

function heroIcon(
  viewBox: string,
  w: number,
  h: number,
  children: VNode[]
): VNode {
  return {
    type: 'svg',
    props: {
      width: w,
      height: h,
      viewBox,
      fill: 'none',
      style: { display: 'block' },
      children,
    },
  };
}

function plusIcon(): VNode {
  return heroIcon('0 0 56 56', 56, 56, [
    { type: 'rect', props: { x: 2, y: 2, width: 52, height: 52, rx: 26, fill: '#AEDFF7' } },
    { type: 'path', props: { d: 'M20 28h16M28 20v16', stroke: '#0F1729', strokeWidth: 2.5, strokeLinecap: 'round' } },
    { type: 'circle', props: { cx: 28, cy: 28, r: 22, stroke: '#0F1729', strokeWidth: 1.5 } },
  ]);
}

function arrowsIcon(): VNode {
  return heroIcon('0 0 64 64', 64, 64, [
    { type: 'rect', props: { x: 2, y: 2, width: 60, height: 60, rx: 30, fill: '#2F6FEB' } },
    { type: 'path', props: { d: 'M22 42l7-7 7 7M36 22l-7 7-7-7', stroke: 'white', strokeWidth: 3, strokeLinecap: 'round', strokeLinejoin: 'round' } },
    { type: 'path', props: { d: 'M29 29v14M35 29v14', stroke: 'white', strokeWidth: 2.5, strokeLinecap: 'round' } },
    { type: 'circle', props: { cx: 32, cy: 32, r: 26, stroke: '#0F1729', strokeWidth: 1.5 } },
  ]);
}

function personIcon(): VNode {
  return heroIcon('0 0 56 56', 56, 56, [
    { type: 'rect', props: { x: 2, y: 2, width: 52, height: 52, rx: 26, fill: '#F5A742' } },
    { type: 'path', props: { d: 'M18 36c0-6 4-10 10-10s10 4 10 10', stroke: '#0F1729', strokeWidth: 2.5, strokeLinecap: 'round' } },
    { type: 'circle', props: { cx: 28, cy: 22, r: 6, stroke: '#0F1729', strokeWidth: 2.5 } },
  ]);
}

export const GET: RequestHandler = async ({ url }) => {
  const parsed = querySchema.safeParse(Object.fromEntries(url.searchParams));

  if (!parsed.success) {
    return new Response('Invalid query params', { status: 400 });
  }

  const { title, description } = parsed.data;

  const [boldData, regularData, antonData, poppinsData] = await Promise.all([
    geistBold, geistRegular, anton, poppinsBold,
  ]);

  const svg = await satori(
    {
      type: 'div',
      props: {
        style: {
          height: '100%',
          width: '100%',
          display: 'flex',
          flexDirection: 'column',
          alignItems: 'center',
          justifyContent: 'center',
          padding: '80px',
          backgroundColor: '#FFFFFF',
          color: '#0F1729',
          fontFamily: 'Geist',
        },
        children: [
          {
            type: 'div',
            props: {
              style: {
                display: 'flex',
                gap: 20,
                marginBottom: 32,
              },
              children: [
                plusIcon(),
                arrowsIcon(),
                personIcon(),
              ],
            },
          },

          {
            type: 'div',
            props: {
              style: {
                fontFamily: 'Poppins',
                fontWeight: 700,
                fontSize: 28,
                color: '#0F1729',
                marginBottom: 8,
              },
              children: 'xians',
            },
          },

          {
            type: 'div',
            props: {
              style: {
                width: 80,
                height: 3,
                backgroundColor: '#7FD8C4',
                borderRadius: 2,
                marginBottom: 24,
              },
              children: null,
            },
          },

          {
            type: 'div',
            props: {
              style: {
                fontFamily: 'Anton',
                fontSize: title === 'xians' ? 72 : (title.length > 40 ? 48 : 64),
                lineHeight: 1.1,
                textTransform: 'uppercase',
                color: '#0F1729',
                textAlign: 'center',
                maxWidth: 900,
                marginBottom: description ? 16 : 0,
              },
              children: title,
            },
          },

          description
            ? {
                type: 'div',
                props: {
                  style: {
                    fontFamily: 'Geist',
                    fontWeight: 400,
                    fontSize: 26,
                    color: '#5B6472',
                    lineHeight: 1.5,
                    textAlign: 'center',
                    maxWidth: 700,
                  },
                  children: description,
                },
              }
            : null,

          {
            type: 'div',
            props: {
              style: {
                display: 'flex',
                alignItems: 'center',
                gap: 6,
                marginTop: 'auto',
                fontFamily: 'Poppins',
                fontWeight: 700,
                fontSize: 15,
                color: '#2F6FEB',
                letterSpacing: 1,
              },
              children: 'xians.vercel.app',
            },
          },
        ].filter(Boolean),
      },
    },
    {
      width: 1200,
      height: 630,
      fonts: [
        { name: 'Geist', data: boldData, weight: 700, style: 'normal' },
        { name: 'Geist', data: regularData, weight: 400, style: 'normal' },
        { name: 'Anton', data: antonData, weight: 400, style: 'normal' },
        { name: 'Poppins', data: poppinsData, weight: 700, style: 'normal' },
      ],
    }
  );

  const resvg = new Resvg(svg, { fitTo: { mode: 'width', value: 1200 } });
  const png = resvg.render().asPng();

  return new Response(png, {
    headers: {
      'Content-Type': 'image/png',
      'Cache-Control': 'public, max-age=86400, s-maxage=604800, stale-while-revalidate=86400',
    },
  });
};
