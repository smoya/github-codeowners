# github-codeowners
A simple application that shows the list of Code Owners from a given list of GH repositories

## Usage

All URIs are relative to *https://github-codeowners.netlify.app*

Method | HTTP request | Description
------------- | ------------- | -------------
[**listCodeOwners**](#listCodeOwners) | **GET** /.netlify/functions/list |


# **listCodeOwners**
    List Code Owners from a given list of GH repositories

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**repo** | **String**| Repository with format: &#x60;&lt;owner&gt;/&lt;repo-name&gt;&#x60;. Note this is a multiple value query param, so you can declare it several times with different values | [default to null]
**format** | **String**| Print format | [optional] [default to null]
**gh\_token** | **String**| Overrides the application GH token. Useful if you need to see private repositories | [optional] [default to null]

### Return type

[**Set**](../Models/inline_response_200.md)

### Authorization

No authorization required

### HTTP request headers

- **Accept**: application/json, text/plain




