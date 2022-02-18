"use strict";(self.webpackChunkaureole=self.webpackChunkaureole||[]).push([[860],{9873:(n,s,e)=>{e.r(s),e.d(s,{data:()=>a});const a={key:"v-df0565f8",path:"/config/crypto_key.html",title:"Крипто-ключ",lang:"en-US",frontmatter:{},excerpt:"",headers:[{level:2,title:"Конфигурация jwk крипто-ключа.",slug:"конфигурация-jwk-крипто-ключа",children:[{level:3,title:"Пример конфига",slug:"пример-конфига",children:[]}]},{level:2,title:"Конфигурация pem крипто-ключа.",slug:"конфигурация-pem-крипто-ключа",children:[{level:3,title:"Пример конфига",slug:"пример-конфига-1",children:[]}]}],filePathRelative:"config/crypto_key.md",git:{contributors:[{name:"Semen Asaevich",email:"semen.asaevich@gmail.com",commits:8},{name:"asaevich",email:"semen.asaevich@gmail.com",commits:1}]}}},4216:(n,s,e)=>{e.r(s),e.d(s,{default:()=>t});const a=(0,e(6252).uE)('<h1 id="крипто-ключ" tabindex="-1"><a class="header-anchor" href="#крипто-ключ" aria-hidden="true">#</a> Крипто-ключ</h1><hr><p>Описывает все свойства данного крипто-ключа. Данный плагин нужен для получения уже существующих, а также генерации ключей, которые в дальнейшем могут быть использованы для подписи данных.</p><ul><li>Тип: <strong>object</strong>. Обязательны: <code>name</code>. Наличие дополнительных свойств: <strong>Недопустимо</strong>. <ul><li><code>name</code>: Тип: <strong>string</strong>. Дополнительно: Имя экземпляра плагина. Имя хранилища. Имя используется для того, чтобы в дальнейшем ссылаться на данный ключ. Минимальная длина: <strong>1</strong>.</li></ul><h4 id="допустим-один-из-вариантов-конфига" tabindex="-1"><a class="header-anchor" href="#допустим-один-из-вариантов-конфига" aria-hidden="true">#</a> Допустим один из вариантов конфига</h4><h2 id="конфигурация-jwk-крипто-ключа" tabindex="-1"><a class="header-anchor" href="#конфигурация-jwk-крипто-ключа" aria-hidden="true">#</a> Конфигурация jwk крипто-ключа.</h2><ul><li>Тип: <strong>object</strong>. Конфигурация jwk крипто-ключа. Служит для задания свойств крипто-ключа. Обязательны: <code>type</code>, <code>config</code>. Наличие дополнительных свойств: <strong>Недопустимо</strong>. <ul><li><code>type</code>: Константа: <strong>jwk</strong>. Тип крипто-ключа. Необходим, чтобы дать Aureole понять, какой из крипто-ключей использовать.</li><li><code>config</code>: Тип: <strong>object</strong>. Конфигурация крипто-ключа. Описывает все свойства данного крипто-ключа. Наличие дополнительных свойств: <strong>Недопустимо</strong>. <h4 id="допустим-один-из-вариантов-конфига-1" tabindex="-1"><a class="header-anchor" href="#допустим-один-из-вариантов-конфига-1" aria-hidden="true">#</a> Допустим один из вариантов конфига</h4><ul><li>Обязательны: <code>storage</code>. <ul><li><code>storage</code>: Тип: <strong>string</strong>. Дополнительно: Ссылка на экземпляр плагина хранилища ключей. Имя хранилища. Хранилище для чтения и записи ключей. Минимальная длина: <strong>1</strong>.</li><li><code>refresh_interval</code>: Тип: <strong>number</strong>. Дополнительно: Единицы измерения: s. Значение по-умолчанию: <strong>86400</strong>. Интервал обновления ключей. Время в секундах, по прошествии которого Aureole будет обновлять ключи. Минимальное значение: <strong>0</strong>.</li><li><code>retries_num</code>: Тип: <strong>integer</strong>. Количество повторений запроса. Максимальное количество повторений запроса, которое будет совершать плагин в случае возникновения ошибки при рефреше ключа. Минимальное значение: <strong>1</strong>.</li><li><code>retry_interval</code>: Тип: <strong>number</strong>. Интервал между запросами. Время, которое будет ждать плагин, перед тем, как совершить очередную попытку сделать запрос на рефреш ключа. Минимальное значение: <strong>0.1</strong>.</li></ul></li><li>Обязательны: <code>storage</code>, <code>kty</code>, <code>alg</code>, <code>size</code>, <code>kid</code>. <ul><li><code>storage</code>: Тип: <strong>string</strong>. Дополнительно: Ссылка на экземпляр плагина хранилища ключей. Имя хранилища. Хранилище для чтения и записи ключей. Минимальная длина: <strong>1</strong>.</li><li><code>refresh_interval</code>: Тип: <strong>number</strong>. Дополнительно: Единицы измерения: s. Значение по-умолчанию: <strong>86400</strong>. Интервал обновления ключей. Время в секундах, по прошествии которого Aureole будет обновлять ключи. Минимальное значение: <strong>0</strong>.</li><li><code>retries_num</code>: Тип: <strong>integer</strong>. Количество повторений запроса. Максимальное количество повторений запроса, которое будет совершать плагин в случае возникновения ошибки при рефреше ключа. Минимальное значение: <strong>1</strong>.</li><li><code>retry_interval</code>: Тип: <strong>number</strong>. Интервал между запросами. Время, которое будет ждать плагин, перед тем, как совершить очередную попытку сделать запрос на рефреш ключа. Минимальное значение: <strong>0.1</strong>.</li><li><code>kty</code>: Допускаются следующие значения: <code>RSA</code>, <code>oct</code>. Тип ключа. Тип семейства криптографических алгоритмов использованных для ключа.</li><li><code>use</code>: Допускаются следующие значения: <code>enc</code>, <code>sig</code>. Предназначение ключа. Определяет, для чего должен использовать ключ&#39;:&#39; подписи или шифрования.</li><li><code>alg</code>: Допускаются следующие значения: <code>RS256</code>, <code>RS384</code>, <code>RS512</code>, <code>RSA-OAEP</code>, <code>RSA-OAEP-256</code>, <code>PS256</code>, <code>PS384</code>, <code>PS512</code>, <code>HS256</code>, <code>HS384</code>, <code>HS512</code>. Алгоритм ключа. Определяет алгоритм, предназначенный для использования с ключом.</li><li><code>size</code>: Тип: <strong>integer</strong>. Размер ключа. Размер генерируемого ключа. Минимальное значение: <strong>512</strong>.</li><li><code>kid</code>: Идентификатор ключа. Метод генерации идентификатора ключа или сам идентификатор. <h4 id="допустим-любои-из-вариантов-конфига" tabindex="-1"><a class="header-anchor" href="#допустим-любои-из-вариантов-конфига" aria-hidden="true">#</a> Допустим любой из вариантов конфига</h4><ul><li>Допускаются следующие значения: <code>SHA-256</code>, <code>SHA-1</code>.</li><li>Тип: <strong>string</strong>. Минимальная длина: <strong>1</strong>.</li></ul></li></ul></li><li>Обязательны: <code>storage</code>, <code>kty</code>, <code>alg</code>, <code>curve</code>, <code>kid</code>. <ul><li><code>storage</code>: Тип: <strong>string</strong>. Дополнительно: Ссылка на экземпляр плагина хранилища ключей. Имя хранилища. Хранилище для чтения и записи ключей. Минимальная длина: <strong>1</strong>.</li><li><code>refresh_interval</code>: Тип: <strong>integer</strong>. Дополнительно: Единицы измерения: ms. Значение по-умолчанию: <strong>3600000</strong>. Интервал обновления ключей. Время в секундах, по прошествии которого Aureole будет обновлять ключи. Минимальное значение: <strong>0</strong>.</li><li><code>retries_num</code>: Тип: <strong>integer</strong>. Количество повторений запроса. Максимальное количество повторений запроса, которое будет совершать плагин в случае возникновения ошибки при рефреше ключа. Минимальное значение: <strong>1</strong>.</li><li><code>retry_interval</code>: Тип: <strong>integer</strong>. Дополнительно: Единицы измерения: ms. Интервал между запросами. Время, которое будет ждать плагин, перед тем, как совершить очередную попытку сделать запрос на рефреш ключа. Минимальное значение: <strong>1</strong>.</li><li><code>kty</code>: Допускаются следующие значения: <code>EC</code>, <code>OKP</code>. Тип ключа. Тип семейства криптографических алгоритмов использованных для ключа.</li><li><code>use</code>: Допускаются следующие значения: <code>enc</code>, <code>sig</code>. Предназначение ключа. Определяет, для чего должен использовать ключ&#39;:&#39; подписи или шифрования.</li><li><code>alg</code>: Допускаются следующие значения: <code>ES256</code>, <code>ES384</code>, <code>ES512</code>, <code>ES256K</code>, <code>EdDSA</code>. Алгоритм ключа. Определяет алгоритм, предназначенный для использования с ключом.</li><li><code>curve</code>: Допускаются следующие значения: <code>P-256</code>, <code>P-384</code>, <code>P-512</code>, <code>Ed25519</code>, <code>Ed448</code>, <code>X25519</code>, <code>X448</code>. Тип кривой. Тип кривой, используемой для генерации ключа.</li><li><code>kid</code>: Идентификатор ключа. Метод генерации идентификатора ключа или сам идентификатор. <h4 id="допустим-любои-из-вариантов-конфига-1" tabindex="-1"><a class="header-anchor" href="#допустим-любои-из-вариантов-конфига-1" aria-hidden="true">#</a> Допустим любой из вариантов конфига</h4><ul><li>Допускаются следующие значения: <code>SHA-256</code>, <code>SHA-1</code>.</li><li>Тип: <strong>string</strong>. Минимальная длина: <strong>1</strong>.</li></ul></li></ul></li></ul></li></ul><h3 id="пример-конфига" tabindex="-1"><a class="header-anchor" href="#пример-конфига" aria-hidden="true">#</a> Пример конфига</h3><div class="language-yaml ext-yml line-numbers-mode"><pre class="language-yaml"><code><span class="token key atrule">crypto_keys</span><span class="token punctuation">:</span>\n  <span class="token comment"># load keys from file</span>\n  <span class="token punctuation">-</span> <span class="token key atrule">type</span><span class="token punctuation">:</span> <span class="token string">&quot;jwk&quot;</span>\n    <span class="token key atrule">name</span><span class="token punctuation">:</span> jwk_file\n    <span class="token key atrule">config</span><span class="token punctuation">:</span>\n      <span class="token key atrule">refresh_interval</span><span class="token punctuation">:</span> <span class="token number">3600</span>\n      <span class="token key atrule">storage</span><span class="token punctuation">:</span> jwk_keys_store\n\n  <span class="token comment"># load keys from url</span>\n  <span class="token punctuation">-</span> <span class="token key atrule">type</span><span class="token punctuation">:</span> <span class="token string">&quot;jwk&quot;</span>\n    <span class="token key atrule">name</span><span class="token punctuation">:</span> jwk_url\n    <span class="token key atrule">config</span><span class="token punctuation">:</span>\n      <span class="token key atrule">refresh_interval</span><span class="token punctuation">:</span> <span class="token number">3600</span>\n      <span class="token key atrule">storage</span><span class="token punctuation">:</span> google_keys_store\n\n  <span class="token comment"># generate keys and save to vault</span>\n  <span class="token punctuation">-</span> <span class="token key atrule">type</span><span class="token punctuation">:</span> <span class="token string">&quot;jwk&quot;</span>\n    <span class="token key atrule">name</span><span class="token punctuation">:</span> jwk_gen_vault\n    <span class="token key atrule">config</span><span class="token punctuation">:</span>\n      <span class="token key atrule">kty</span><span class="token punctuation">:</span> RSA\n      <span class="token key atrule">alg</span><span class="token punctuation">:</span> RS256\n      <span class="token key atrule">size</span><span class="token punctuation">:</span> <span class="token number">2048</span>\n      <span class="token key atrule">kid</span><span class="token punctuation">:</span> SHA<span class="token punctuation">-</span><span class="token number">256</span>\n      <span class="token key atrule">refresh_interval</span><span class="token punctuation">:</span> <span class="token number">3600</span>\n      <span class="token key atrule">storage</span><span class="token punctuation">:</span> vault_keys_store\n</code></pre><div class="line-numbers"><span class="line-number">1</span><br><span class="line-number">2</span><br><span class="line-number">3</span><br><span class="line-number">4</span><br><span class="line-number">5</span><br><span class="line-number">6</span><br><span class="line-number">7</span><br><span class="line-number">8</span><br><span class="line-number">9</span><br><span class="line-number">10</span><br><span class="line-number">11</span><br><span class="line-number">12</span><br><span class="line-number">13</span><br><span class="line-number">14</span><br><span class="line-number">15</span><br><span class="line-number">16</span><br><span class="line-number">17</span><br><span class="line-number">18</span><br><span class="line-number">19</span><br><span class="line-number">20</span><br><span class="line-number">21</span><br><span class="line-number">22</span><br><span class="line-number">23</span><br><span class="line-number">24</span><br><span class="line-number">25</span><br></div></div></li></ul><h2 id="конфигурация-pem-крипто-ключа" tabindex="-1"><a class="header-anchor" href="#конфигурация-pem-крипто-ключа" aria-hidden="true">#</a> Конфигурация pem крипто-ключа.</h2><ul><li>Тип: <strong>object</strong>. Конфигурация pem крипто-ключа. Служит для задания свойств крипто-ключа. Обязательны: <code>type</code>, <code>config</code>. Наличие дополнительных свойств: <strong>Недопустимо</strong>. <ul><li><code>type</code>: Константа: <strong>pem</strong>. Тип крипто-ключа. Необходим, чтобы дать Aureole понять, какой из крипто-ключей использовать.</li><li><code>config</code>: Тип: <strong>object</strong>. Конфигурация крипто-ключа. Описывает все свойства данного крипто-ключа. Обязательны: <code>alg</code>, <code>storage</code>. Наличие дополнительных свойств: <strong>Недопустимо</strong>. <ul><li><code>alg</code>: Тип: <strong>string</strong>. Алгоритм ключа. Описывает алгоритм ключа, хранящегося по данному пути. Минимальная длина: <strong>1</strong>.</li><li><code>storage</code>: Тип: <strong>string</strong>. Дополнительно: Ссылка на экземпляр плагина хранилища ключей. Имя хранилища. Хранилище для чтения и записи ключей. Минимальная длина: <strong>1</strong>.</li><li><code>refresh_interval</code>: Тип: <strong>integer</strong>. Дополнительно: Единицы измерения: ms. Значение по-умолчанию: <strong>3600000</strong>. Интервал обновления ключей. Время в секундах, по прошествии которого Aureole будет обновлять ключи. Минимальное значение: <strong>0</strong>.</li><li><code>retries_num</code>: Тип: <strong>integer</strong>. Количество повторений запроса. Максимальное количество повторений запроса, которое будет совершать плагин в случае возникновения ошибки при рефреше ключа. Минимальное значение: <strong>1</strong>.</li><li><code>retry_interval</code>: Тип: <strong>integer</strong>. Дополнительно: Единицы измерения: ms. Интервал между запросами. Время, которое будет ждать плагин, перед тем, как совершить очередную попытку сделать запрос на рефреш ключа. Минимальное значение: <strong>1</strong>.</li></ul></li></ul><h3 id="пример-конфига-1" tabindex="-1"><a class="header-anchor" href="#пример-конфига-1" aria-hidden="true">#</a> Пример конфига</h3><div class="language-yaml ext-yml line-numbers-mode"><pre class="language-yaml"><code><span class="token key atrule">crypto_keys</span><span class="token punctuation">:</span>\n  <span class="token punctuation">-</span> <span class="token key atrule">type</span><span class="token punctuation">:</span> <span class="token string">&quot;pem&quot;</span>\n    <span class="token key atrule">name</span><span class="token punctuation">:</span> pem_key\n    <span class="token key atrule">config</span><span class="token punctuation">:</span>\n      <span class="token key atrule">alg</span><span class="token punctuation">:</span> ES256\n      <span class="token key atrule">refresh_interval</span><span class="token punctuation">:</span> <span class="token number">3600</span>\n      <span class="token key atrule">storage</span><span class="token punctuation">:</span> pem_keys_store\n</code></pre><div class="line-numbers"><span class="line-number">1</span><br><span class="line-number">2</span><br><span class="line-number">3</span><br><span class="line-number">4</span><br><span class="line-number">5</span><br><span class="line-number">6</span><br><span class="line-number">7</span><br></div></div></li></ul></li></ul>',4),o={},t=(0,e(3744).Z)(o,[["render",function(n,s){return a}]])},3744:(n,s)=>{s.Z=(n,s)=>{for(const[e,a]of s)n[e]=a;return n}}}]);