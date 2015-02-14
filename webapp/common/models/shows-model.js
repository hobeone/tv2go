showService = angular.module('tv2go.showsService', ['ngResource']);

showService.factory('Show', ['$resource',
  function($resource) {
    return $resource('data/:showId.json', {}, {
      update: {
        method: "PUT"
      },
      query: {
        method: "GET",
        params: {
          showId: 'shows',
        },
        isArray:true
      }
    });
  }]);

angular.module('tv2go.models.shows',['tv2go.showsService'])
.service('ShowsModel', function($http, $q, Show) {
  var model = this;
  var URLS = {
    FETCH: "data/shows.json",
  };
  var shows;
  var currentShow;

  function cacheShows(result) {
    console.log(result);
    shows = result;
    return shows;
  }
  model.getShows = function() {
    return (shows) ? $q.when(shows): Show.query().$promise.then(cacheShows);
  };

  model.setCurrentShow = function(showName) {
    return model.getShowByName(showName)
    .then(function(show){
      currentShow = show;
    });
  };

  model.getCurrentShow = function() {
    return currentShow;
  };

  model.getCurrentShowName = function() {
    return currentShow ? currentShow.name : "";
  };

  model.getShowByName = function(showName) {
    var deferred = $q.defer();
    function findShow() {
      return _.find(shows, function(s) {
        return s.name == showName;
      });
    }
    if(shows) {
      deferred.resolve(findShow());
    } else {
      model.getShows()
      .then(function(result){
        deferred.resolve(findShow());
      });
    }
    return deferred.promise;
  };
})
;
