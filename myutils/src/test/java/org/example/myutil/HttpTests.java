package org.example.myutil;import org.example.myutil.utils.HttpClientUtils;import org.example.myutil.utils.JsonUtils;import org.junit.jupiter.api.Assertions;import org.junit.jupiter.api.Test;import java.io.File;import java.io.IOException;import java.io.InputStream;import java.net.http.HttpResponse;import java.nio.charset.StandardCharsets;import java.util.HashMap;import java.util.Map;import java.util.Optional;import java.util.logging.Level;import java.util.logging.Logger;/** * @author hzhq1255 * @version 1.0 * @since 2022-11-27 下午5:09 */public class HttpTests {    private static final String BASE_URL = "http://localhost:6443";    private static final Logger LOGGER = Logger.getLogger("HttpTests");    @Test    public void doPostYaml() throws IOException, InterruptedException {        String url = BASE_URL + "/api/v1/namespaces/default/pods";        InputStream podInputStream = getClass().getResourceAsStream("/Pod.yaml");        byte[] bytes = podInputStream == null ? new byte[]{} : podInputStream.readAllBytes();        String podYaml = new String(bytes, StandardCharsets.UTF_8);        Map<String, String> headers = new HashMap<>();        headers.put("content-type", "application/yaml");        HttpResponse<String> response = HttpClientUtils.doPost(url, headers, new HashMap<>(){{            put("pretty", "true");        }}, podYaml);        Assertions.assertNotEquals(null, response);        LOGGER.log(Level.INFO, String.format("response body: %s", response.body()));        Assertions.assertEquals(201, response.statusCode());;        Thread.sleep(2000);    }    @Test    public void doGet() throws IOException, InterruptedException {        String url = BASE_URL + "/api/v1/namespaces/default/pods/myapp";        HttpResponse<String> response = HttpClientUtils.doGet(url, null, new HashMap<>() {{            put("pretty", "true");        }});        Assertions.assertNotEquals(null, response);        LOGGER.log(Level.INFO, String.format("response body: %s", response.body()));        Assertions.assertEquals(200, response.statusCode());        Assertions.assertTrue(JsonUtils.isJSONValid(response.body()));        Thread.sleep(2000);    }    @Test    public void doPutYaml() throws IOException, InterruptedException {        String url = BASE_URL + "/api/v1/namespaces/default/pods/myapp";        InputStream podInputStream = getClass().getResourceAsStream("/PodUpdate.yaml");        byte[] bytes = podInputStream == null ? new byte[]{} : podInputStream.readAllBytes();        String podYaml = new String(bytes, StandardCharsets.UTF_8);        Map<String, String> headers = new HashMap<>();        headers.put("content-type", "application/yaml");        HttpResponse<String> response = HttpClientUtils.doPut(url, headers, new HashMap<>() {{            put("pretty", "true");        }}, podYaml);        Assertions.assertNotEquals(null, response);        LOGGER.log(Level.INFO, String.format("response body: %s", response.body()));        Assertions.assertEquals(200, response.statusCode());        Thread.sleep(2000);    }    @Test    public void doDelete() throws IOException, InterruptedException {        String url = BASE_URL + "/api/v1/namespaces/default/pods/myapp";        HttpResponse<String> response = HttpClientUtils.doDelete(url, null, new HashMap<>() {{            put("pretty", "true");        }});        Assertions.assertNotEquals(null, response);        LOGGER.log(Level.INFO, String.format("response body: %s", response.body()));        Assertions.assertEquals(200, response.statusCode());        Thread.sleep(2000);    }}