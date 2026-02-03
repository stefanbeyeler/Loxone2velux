import { useState } from 'react';
import {
  ChevronDown,
  ChevronRight,
  Copy,
  Check,
} from 'lucide-react';

interface CodeBlockProps {
  code: string;
  language?: string;
}

function CodeBlock({ code }: CodeBlockProps) {
  const [copied, setCopied] = useState(false);

  const handleCopy = async () => {
    await navigator.clipboard.writeText(code);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <div className="relative group">
      <pre className="bg-gray-900 rounded-lg p-4 overflow-x-auto text-sm">
        <code className="text-gray-300">{code}</code>
      </pre>
      <button
        onClick={handleCopy}
        className="absolute top-2 right-2 p-2 bg-gray-700 hover:bg-gray-600 rounded-lg opacity-0 group-hover:opacity-100 transition-opacity"
      >
        {copied ? (
          <Check size={16} className="text-green-400" />
        ) : (
          <Copy size={16} className="text-gray-400" />
        )}
      </button>
    </div>
  );
}

interface AccordionProps {
  title: string;
  children: React.ReactNode;
  defaultOpen?: boolean;
}

function Accordion({ title, children, defaultOpen = false }: AccordionProps) {
  const [isOpen, setIsOpen] = useState(defaultOpen);

  return (
    <div className="border border-gray-700 rounded-lg overflow-hidden">
      <button
        onClick={() => setIsOpen(!isOpen)}
        className="w-full flex items-center justify-between p-4 bg-gray-800 hover:bg-gray-750 transition-colors"
      >
        <span className="font-medium text-white">{title}</span>
        {isOpen ? (
          <ChevronDown size={20} className="text-gray-400" />
        ) : (
          <ChevronRight size={20} className="text-gray-400" />
        )}
      </button>
      {isOpen && (
        <div className="p-4 bg-gray-800/50 border-t border-gray-700">
          {children}
        </div>
      )}
    </div>
  );
}

export function LoxoneGuide() {
  const token = localStorage.getItem('api_token') || 'DEIN_TOKEN';

  return (
    <div className="space-y-6">
      <div className="bg-gray-800 rounded-xl p-6">
        <h2 className="text-xl font-bold text-white mb-2">Loxone Integration</h2>
        <p className="text-gray-400">
          Anleitung zur Einbindung der Velux-Geräte in Loxone über Virtual HTTP Outputs.
        </p>
      </div>

      <div className="space-y-4">
        <Accordion title="1. Virtual HTTP Output erstellen" defaultOpen>
          <div className="space-y-4 text-gray-300">
            <p>
              Erstelle in Loxone Config einen neuen <strong>Virtual HTTP Output</strong>:
            </p>
            <ol className="list-decimal list-inside space-y-2 ml-4">
              <li>Rechtsklick auf "Virtual Outputs" → "Virtuellen HTTP Ausgang hinzufügen"</li>
              <li>Adresse des Gateways eintragen (z.B. <code className="bg-gray-900 px-2 py-0.5 rounded">http://192.168.1.100:8080</code>)</li>
              <li>Keine Authentifizierung konfigurieren (Token wird per URL übergeben)</li>
            </ol>
          </div>
        </Accordion>

        <Accordion title="2. Virtual Output Commands">
          <div className="space-y-4 text-gray-300">
            <p>Füge die folgenden Befehle zum Virtual HTTP Output hinzu:</p>

            <div className="space-y-4">
              <div>
                <h4 className="font-medium text-white mb-2">Position setzen (0-100%)</h4>
                <CodeBlock code={`/loxone/node/<ID>/set/<v>?token=${token}`} />
                <p className="text-sm text-gray-400 mt-2">
                  Ersetze <code>&lt;ID&gt;</code> mit der Geräte-ID und <code>&lt;v&gt;</code> mit dem Analogwert (0-100).
                </p>
              </div>

              <div>
                <h4 className="font-medium text-white mb-2">Öffnen</h4>
                <CodeBlock code={`/loxone/node/<ID>/open?token=${token}`} />
              </div>

              <div>
                <h4 className="font-medium text-white mb-2">Schließen</h4>
                <CodeBlock code={`/loxone/node/<ID>/close?token=${token}`} />
              </div>

              <div>
                <h4 className="font-medium text-white mb-2">Stoppen</h4>
                <CodeBlock code={`/loxone/node/<ID>/stop?token=${token}`} />
              </div>
            </div>
          </div>
        </Accordion>

        <Accordion title="3. Beispiel: Fenster-Baustein">
          <div className="space-y-4 text-gray-300">
            <p>
              So verbindest du einen Jalousie-Baustein mit dem Gateway:
            </p>
            <ol className="list-decimal list-inside space-y-2 ml-4">
              <li>Erstelle einen <strong>Jalousie</strong> oder <strong>Beschattung</strong> Baustein</li>
              <li>Verbinde die Ausgänge mit dem Virtual HTTP Output:</li>
            </ol>

            <div className="mt-4 bg-gray-900 rounded-lg p-4">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b border-gray-700">
                    <th className="text-left py-2 text-gray-400">Ausgang</th>
                    <th className="text-left py-2 text-gray-400">HTTP Befehl</th>
                  </tr>
                </thead>
                <tbody className="font-mono">
                  <tr className="border-b border-gray-800">
                    <td className="py-2">AQ (Position)</td>
                    <td className="py-2 text-velux-blue">/loxone/node/1/set/&lt;v&gt;?token=...</td>
                  </tr>
                  <tr className="border-b border-gray-800">
                    <td className="py-2">QU (Hoch)</td>
                    <td className="py-2 text-velux-blue">/loxone/node/1/open?token=...</td>
                  </tr>
                  <tr className="border-b border-gray-800">
                    <td className="py-2">QD (Runter)</td>
                    <td className="py-2 text-velux-blue">/loxone/node/1/close?token=...</td>
                  </tr>
                  <tr>
                    <td className="py-2">QS (Stopp)</td>
                    <td className="py-2 text-velux-blue">/loxone/node/1/stop?token=...</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </Accordion>

        <Accordion title="4. REST API Referenz">
          <div className="space-y-4 text-gray-300">
            <p>Vollständige API-Endpunkte:</p>

            <div className="bg-gray-900 rounded-lg p-4 overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b border-gray-700">
                    <th className="text-left py-2 text-gray-400">Methode</th>
                    <th className="text-left py-2 text-gray-400">Endpoint</th>
                    <th className="text-left py-2 text-gray-400">Beschreibung</th>
                  </tr>
                </thead>
                <tbody className="font-mono">
                  <tr className="border-b border-gray-800">
                    <td className="py-2 text-green-400">GET</td>
                    <td className="py-2">/health</td>
                    <td className="py-2 font-sans text-gray-400">Status (ohne Auth)</td>
                  </tr>
                  <tr className="border-b border-gray-800">
                    <td className="py-2 text-green-400">GET</td>
                    <td className="py-2">/api/nodes</td>
                    <td className="py-2 font-sans text-gray-400">Alle Geräte</td>
                  </tr>
                  <tr className="border-b border-gray-800">
                    <td className="py-2 text-green-400">GET</td>
                    <td className="py-2">/api/nodes/&#123;id&#125;</td>
                    <td className="py-2 font-sans text-gray-400">Einzelnes Gerät</td>
                  </tr>
                  <tr className="border-b border-gray-800">
                    <td className="py-2 text-yellow-400">POST</td>
                    <td className="py-2">/api/nodes/&#123;id&#125;/position</td>
                    <td className="py-2 font-sans text-gray-400">Position setzen</td>
                  </tr>
                  <tr className="border-b border-gray-800">
                    <td className="py-2 text-yellow-400">POST</td>
                    <td className="py-2">/api/nodes/&#123;id&#125;/open</td>
                    <td className="py-2 font-sans text-gray-400">Öffnen</td>
                  </tr>
                  <tr className="border-b border-gray-800">
                    <td className="py-2 text-yellow-400">POST</td>
                    <td className="py-2">/api/nodes/&#123;id&#125;/close</td>
                    <td className="py-2 font-sans text-gray-400">Schließen</td>
                  </tr>
                  <tr>
                    <td className="py-2 text-yellow-400">POST</td>
                    <td className="py-2">/api/nodes/&#123;id&#125;/stop</td>
                    <td className="py-2 font-sans text-gray-400">Stoppen</td>
                  </tr>
                </tbody>
              </table>
            </div>

            <p className="text-sm text-gray-400">
              Authentifizierung: <code className="bg-gray-900 px-2 py-0.5 rounded">Authorization: Bearer &#123;token&#125;</code> oder <code className="bg-gray-900 px-2 py-0.5 rounded">?token=&#123;token&#125;</code>
            </p>
          </div>
        </Accordion>

        <Accordion title="5. Troubleshooting">
          <div className="space-y-4 text-gray-300">
            <div>
              <h4 className="font-medium text-white mb-2">Gateway nicht erreichbar</h4>
              <ul className="list-disc list-inside space-y-1 ml-4 text-gray-400">
                <li>Prüfe ob der Docker Container läuft: <code>docker compose ps</code></li>
                <li>Prüfe die Logs: <code>docker compose logs -f</code></li>
                <li>Stelle sicher, dass Port 8080 erreichbar ist</li>
              </ul>
            </div>

            <div>
              <h4 className="font-medium text-white mb-2">KLF-200 nicht verbunden</h4>
              <ul className="list-disc list-inside space-y-1 ml-4 text-gray-400">
                <li>Überprüfe IP-Adresse und Passwort in config.yaml</li>
                <li>Das Passwort steht auf der Rückseite des KLF-200</li>
                <li>Stelle sicher, dass der KLF-200 im selben Netzwerk ist</li>
              </ul>
            </div>

            <div>
              <h4 className="font-medium text-white mb-2">401 Unauthorized</h4>
              <ul className="list-disc list-inside space-y-1 ml-4 text-gray-400">
                <li>API Token prüfen (min. 16 Zeichen)</li>
                <li>Token korrekt in URL oder Header übergeben</li>
              </ul>
            </div>
          </div>
        </Accordion>
      </div>
    </div>
  );
}
