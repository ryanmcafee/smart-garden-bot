import type { Metadata, Viewport } from 'next';
import { Inter } from 'next/font/google';
import { Providers } from './providers';
import { Toaster } from '@/components/ui/toaster';
import { ThemeProvider } from '@/components/theme-provider';
import './globals.css';

const inter = Inter({ subsets: ['latin'] });

export const metadata: Metadata = {
  title: {
    template: '%s | Smart Garden Bot',
    default: 'Smart Garden Bot - Intelligent Garden Automation',
  },
  description: 'Automate your garden watering with intelligent weather-based decisions and IoT sensor integration.',
  keywords: ['garden', 'automation', 'IoT', 'watering', 'smart home', 'agriculture'],
  authors: [{ name: 'Smart Garden Bot Team' }],
  creator: 'Smart Garden Bot',
  publisher: 'Smart Garden Bot',
  robots: {
    index: true,
    follow: true,
    googleBot: {
      index: true,
      follow: true,
      'max-video-preview': -1,
      'max-image-preview': 'large',
      'max-snippet': -1,
    },
  },
  openGraph: {
    type: 'website',
    locale: 'en_US',
    url: 'https://smartgardenbot.com',
    title: 'Smart Garden Bot - Intelligent Garden Automation',
    description: 'Automate your garden watering with intelligent weather-based decisions and IoT sensor integration.',
    siteName: 'Smart Garden Bot',
    images: [
      {
        url: '/og-image.jpg',
        width: 1200,
        height: 630,
        alt: 'Smart Garden Bot',
      },
    ],
  },
  twitter: {
    card: 'summary_large_image',
    title: 'Smart Garden Bot - Intelligent Garden Automation',
    description: 'Automate your garden watering with intelligent weather-based decisions and IoT sensor integration.',
    images: ['/og-image.jpg'],
    creator: '@smartgardenbot',
  },
  manifest: '/manifest.json',
  icons: {
    icon: '/favicon.ico',
    shortcut: '/favicon-16x16.png',
    apple: '/apple-touch-icon.png',
  },
};

export const viewport: Viewport = {
  width: 'device-width',
  initialScale: 1,
  maximumScale: 1,
  userScalable: false,
  themeColor: [
    { media: '(prefers-color-scheme: light)', color: '#ffffff' },
    { media: '(prefers-color-scheme: dark)', color: '#000000' },
  ],
};

interface RootLayoutProps {
  children: React.ReactNode;
}

export default function RootLayout({ children }: RootLayoutProps) {
  return (
    <html lang="en" suppressHydrationWarning>
      <body className={inter.className}>
        <ThemeProvider
          attribute="class"
          defaultTheme="system"
          enableSystem
          disableTransitionOnChange
        >
          <Providers>
            <div className="min-h-screen bg-background">
              {children}
            </div>
            <Toaster />
          </Providers>
        </ThemeProvider>
      </body>
    </html>
  );
}